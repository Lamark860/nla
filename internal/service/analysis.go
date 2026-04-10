package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nla/internal/client/openai"
	"nla/internal/model"
	"nla/internal/mongo"
)

type AnalysisService struct {
	repo       *mongo.AnalysisRepo
	openai     *openai.Client
	promptPath string
}

func NewAnalysisService(repo *mongo.AnalysisRepo, openaiClient *openai.Client, promptPath string) *AnalysisService {
	return &AnalysisService{
		repo:       repo,
		openai:     openaiClient,
		promptPath: promptPath,
	}
}

// Analyze runs AI analysis for a bond and returns the result
func (s *AnalysisService) Analyze(ctx context.Context, secid string, bondDataJSON string) (*model.BondAnalysis, error) {
	prompt, err := s.loadPrompt()
	if err != nil {
		return nil, fmt.Errorf("load prompt: %w", err)
	}

	cleanedJSON := prepareJSONForAI(bondDataJSON)
	userMessage := fmt.Sprintf(
		"Проанализируй следующую облигацию.\n\nДата анализа: %s\n\nДанные:\n%s",
		time.Now().Format("2006-01-02"),
		cleanedJSON,
	)

	messages := []openai.ChatMessage{
		{Role: "system", Content: prompt},
		{Role: "user", Content: userMessage},
	}

	response, err := s.openai.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openai completion: %w", err)
	}

	rating := ParseRatingFromResponse(response)

	analysis := &model.BondAnalysis{
		ID:        generateID("analysis_"),
		SECID:     secid,
		Response:  response,
		Rating:    rating,
		Timestamp: time.Now(),
	}

	if err := s.repo.Save(ctx, analysis); err != nil {
		return nil, fmt.Errorf("save analysis: %w", err)
	}

	return analysis, nil
}

// GetBySecid returns all analyses for a bond
func (s *AnalysisService) GetBySecid(ctx context.Context, secid string) ([]model.BondAnalysis, error) {
	return s.repo.GetBySecid(ctx, secid)
}

// GetByID returns a single analysis
func (s *AnalysisService) GetByID(ctx context.Context, id string) (*model.BondAnalysis, error) {
	return s.repo.GetByID(ctx, id)
}

// GetStats returns aggregate stats for a bond
func (s *AnalysisService) GetStats(ctx context.Context, secid string) (*model.AnalysisStats, error) {
	return s.repo.GetStats(ctx, secid)
}

// GetLatestRatings batch-loads latest ratings for multiple SECIDs
func (s *AnalysisService) GetLatestRatings(ctx context.Context, secids []string) (map[string]int, error) {
	return s.repo.GetLatestRatings(ctx, secids)
}

// GetBulkStats returns analysis stats for all analyzed bonds
func (s *AnalysisService) GetBulkStats(ctx context.Context) (map[string]model.AnalysisStats, error) {
	return s.repo.GetBulkStats(ctx)
}

func (s *AnalysisService) loadPrompt() (string, error) {
	data, err := os.ReadFile(s.promptPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseRatingFromResponse extracts rating 0-100 from AI response text.
// Priority: [RATING:XX] > "Итоговая оценка" XX/100 > "Итоговая" XX баллов >
// "Оценка" XX/100 > "Оценка" XX баллов > **XX/100** > any XX/100
func ParseRatingFromResponse(text string) *int {
	// 1. Machine-readable tag [RATING:XX]
	re1 := regexp.MustCompile(`\[RATING:\s*(\d{1,3})\s*\]`)
	if m := re1.FindStringSubmatch(text); m != nil {
		if v := parseAndValidate(m[1]); v != nil {
			return v
		}
	}

	// 2. "Итоговая оценка" block with XX/100
	re2 := regexp.MustCompile(`(?i)[ии]тогов\w*\s+оценк\w*[^\n]{0,80}?(\d{1,3})\s*[/\\]\s*100`)
	if m := re2.FindStringSubmatch(text); m != nil {
		if v := parseAndValidate(m[1]); v != nil {
			return v
		}
	}

	// 3. "Итоговая оценка: XX баллов"
	re3 := regexp.MustCompile(`(?i)[ии]тогов\w*\s+оценк\w*[^\n]{0,80}?(\d{1,3})\s*балл`)
	if m := re3.FindStringSubmatch(text); m != nil {
		if v := parseAndValidate(m[1]); v != nil {
			return v
		}
	}

	// 4. "Оценка: XX/100" (without "Итоговая")
	re4 := regexp.MustCompile(`(?i)[оо]ценк\w*\s*:?\s*[^\n]{0,40}?(\d{1,3})\s*[/\\]\s*100`)
	if m := re4.FindStringSubmatch(text); m != nil {
		if v := parseAndValidate(m[1]); v != nil {
			return v
		}
	}

	// 5. "Оценка: XX баллов"
	re5 := regexp.MustCompile(`(?i)[оо]ценк\w*\s*:?\s*[^\n]{0,40}?(\d{1,3})\s*балл`)
	if m := re5.FindStringSubmatch(text); m != nil {
		if v := parseAndValidate(m[1]); v != nil {
			return v
		}
	}

	// 6. Any **XX/100** (bold markdown)
	re6 := regexp.MustCompile(`\*\*\s*(\d{1,3})\s*[/\\]\s*100\s*\*\*`)
	if m := re6.FindStringSubmatch(text); m != nil {
		if v := parseAndValidate(m[1]); v != nil {
			return v
		}
	}

	// 7. Any XX/100 (last occurrence, least reliable)
	re7 := regexp.MustCompile(`(\d{1,3})\s*[/\\]\s*100`)
	matches := re7.FindAllStringSubmatch(text, -1)
	if len(matches) > 0 {
		last := matches[len(matches)-1]
		if v := parseAndValidate(last[1]); v != nil {
			return v
		}
	}

	return nil
}

func parseAndValidate(s string) *int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	if v >= 0 && v <= 100 {
		return &v
	}
	return nil
}

// prepareJSONForAI cleans and compresses bond data before sending to AI
func prepareJSONForAI(jsonStr string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr
	}

	// Remove duplicate security/marketdata in coupons (already in bond)
	if coupons, ok := data["coupons"].(map[string]any); ok {
		if _, hasBond := data["bond"]; hasBond {
			delete(coupons, "security")
			delete(coupons, "marketdata")
			delete(coupons, "marketdata_yields")
		}
	}

	// Limit history to last 15 candles
	if history, ok := data["history"].(map[string]any); ok {
		if candles, ok := history["candles"].(map[string]any); ok {
			if candleData, ok := candles["data"].([]any); ok && len(candleData) > 15 {
				candles["data"] = candleData[len(candleData)-15:]
			}
			delete(candles, "columns")
		}
	}

	// Recursively clean: round floats, remove nulls
	data = cleanDataForAI(data)

	result, err := json.Marshal(data)
	if err != nil {
		return jsonStr
	}
	return string(result)
}

func cleanDataForAI(data map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range data {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case map[string]any:
			cleaned := cleanDataForAI(val)
			if len(cleaned) > 0 {
				result[k] = cleaned
			}
		case []any:
			cleaned := cleanSliceForAI(val)
			if len(cleaned) > 0 {
				result[k] = cleaned
			}
		case float64:
			abs := math.Abs(val)
			if abs > 10000 {
				result[k] = math.Round(val)
			} else if abs > 100 {
				result[k] = math.Round(val*10) / 10
			} else {
				result[k] = math.Round(val*100) / 100
			}
		default:
			result[k] = v
		}
	}
	return result
}

func cleanSliceForAI(data []any) []any {
	var result []any
	for _, item := range data {
		if item == nil {
			continue
		}
		if m, ok := item.(map[string]any); ok {
			cleaned := cleanDataForAI(m)
			if len(cleaned) > 0 {
				result = append(result, cleaned)
			}
		} else {
			result = append(result, item)
		}
	}
	return result
}

func generateID(prefix string) string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		log.Printf("WARN: rand.Read failed: %v, using timestamp", err)
		return prefix + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return prefix + hex.EncodeToString(b[:8]) + "-" + hex.EncodeToString(b[8:])
}

// GenerateChatCompletion sends messages with system prompt to OpenAI (for chat)
func (s *AnalysisService) GenerateChatCompletion(ctx context.Context, messages []openai.ChatMessage, systemPrompt string) (string, error) {
	allMessages := make([]openai.ChatMessage, 0, len(messages)+1)
	allMessages = append(allMessages, openai.ChatMessage{Role: "system", Content: systemPrompt})
	allMessages = append(allMessages, messages...)

	resp, err := s.openai.ChatCompletion(ctx, allMessages)
	if err != nil {
		return "", fmt.Errorf("chat completion: %w", err)
	}
	return resp, nil
}

// PromptForAgent loads a prompt file by agent type
func PromptForAgent(dataDir, agentType string) (string, error) {
	path := strings.TrimRight(dataDir, "/") + "/" + agentType + ".txt"
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load agent prompt %q: %w", agentType, err)
	}
	return string(data), nil
}
