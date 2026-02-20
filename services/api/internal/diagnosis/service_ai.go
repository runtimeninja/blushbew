package diagnosis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/runtimeninja/blushbew/services/api/internal/ai"
)

type Service struct {
	ai *ai.Client
}

func NewService(aiClient *ai.Client) *Service {
	return &Service{ai: aiClient}
}

func (s *Service) Analyze(ctx context.Context, in Input) (Output, error) {
	// Always have a safe fallback
	fb := Fallback(in)

	if s.ai == nil || !s.ai.IsEnabled() {
		return fb, nil
	}

	system := `You are a beauty troubleshooting assistant.
Rules:
- DO NOT provide medical diagnosis.
- Keep advice practical, safe, and generic.
- Output MUST be valid JSON and match the schema exactly.
- No markdown, no extra keys.

Schema:
{
  "likely_causes": ["..."],
  "fix_steps": ["..."],
  "do": ["..."],
  "dont": ["..."],
  "suggested_product_types": ["..."]
}

Constraints:
- Each array: 3 to 6 items.
- Keep items short, actionable, and not brand-specific.
- If user mentions irritation, recommend patch test and stop use if irritation worsens.`

	user := buildUserPrompt(in)

	// small retry once for transient issues
	outStr, err := s.ai.GenerateJSON(ctx, system, user)
	if err != nil {
		return fb, nil
	}

	// Clean common “```json ...```” wrappers if model returns them
	outStr = strings.TrimSpace(outStr)
	outStr = strings.TrimPrefix(outStr, "```json")
	outStr = strings.TrimPrefix(outStr, "```")
	outStr = strings.TrimSuffix(outStr, "```")
	outStr = strings.TrimSpace(outStr)

	var out Output
	if err := json.Unmarshal([]byte(outStr), &out); err != nil {
		return fb, nil
	}

	if !validOutput(out) {
		return fb, nil
	}

	return out, nil
}

func buildUserPrompt(in Input) string {
	now := time.Now().Format(time.RFC3339)
	return fmt.Sprintf(
		`Time: %s
User problem: %s
Optional context:
- Skin type: %s
- Climate: %s
- Event: %s

Give troubleshooting guidance for why the makeup breaks and how to fix it.`,
		now,
		strings.TrimSpace(in.ProblemText),
		strings.TrimSpace(in.SkinType),
		strings.TrimSpace(in.Climate),
		strings.TrimSpace(in.EventType),
	)
}

func validOutput(o Output) bool {
	return len(o.LikelyCauses) >= 3 &&
		len(o.FixSteps) >= 3 &&
		len(o.Do) >= 3 &&
		len(o.Dont) >= 3 &&
		len(o.SuggestedProductTypes) >= 3
}
