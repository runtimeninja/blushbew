package diagnosis

import (
	"strings"
)

type Input struct {
	ProblemText string `json:"problem_text"`
	SkinType    string `json:"skin_type,omitempty"`
	Climate     string `json:"climate,omitempty"`
	EventType   string `json:"event_type,omitempty"`
}

type Output struct {
	LikelyCauses          []string `json:"likely_causes"`
	FixSteps              []string `json:"fix_steps"`
	Do                    []string `json:"do"`
	Dont                  []string `json:"dont"`
	SuggestedProductTypes []string `json:"suggested_product_types"`
}

func Fallback(in Input) Output {
	t := strings.ToLower(in.ProblemText)

	out := Output{
		LikelyCauses: []string{
			"Skin prep mismatch (hydration / oil control imbalance)",
			"Primer + foundation base conflict (water vs silicone)",
			"Not setting properly (powder/spray or wrong order)",
		},
		FixSteps: []string{
			"Clean base → light moisturizer → wait 3–5 min",
			"Choose compatible primer + foundation base (both water-based or both silicone-based)",
			"Apply thin layers, let each layer set",
			"Use setting powder on T-zone then setting spray (optional)",
		},
		Do: []string{
			"Patch test new products",
			"Blot excess oil before reapplying",
		},
		Dont: []string{
			"Mix heavy skincare with matte base immediately",
			"Over-layer thick foundation (it cracks/melts faster)",
		},
		SuggestedProductTypes: []string{
			"Oil-control primer (if oily/humid)",
			"Hydrating primer (if dry/flaky)",
			"Setting powder (translucent)",
			"Setting spray (long-wear)",
		},
	}

	if strings.Contains(t, "crack") || strings.Contains(t, "cak") {
		out.LikelyCauses = append([]string{"Too much product + dry patches causing cracking/caking"}, out.LikelyCauses...)
		out.FixSteps = append([]string{"Exfoliate gently (1–2x/week) + hydrate; avoid piling base on dry areas"}, out.FixSteps...)
	}
	if strings.Contains(t, "melt") || strings.Contains(t, "slip") {
		out.LikelyCauses = append([]string{"High humidity + excess oil causing makeup to slip/melt"}, out.LikelyCauses...)
		out.FixSteps = append([]string{"Use matte/oil-control base for humid weather; set with powder"}, out.FixSteps...)
	}

	return out
}
