package llm

import "katip/internal/hardware"

func EnrichCatalogWithFit(profile hardware.Profile) []ModelInfo {
	result := make([]ModelInfo, len(ModelCatalog))
	for i, m := range ModelCatalog {
		result[i] = m
		result[i].FitStatus = hardware.ComputeFitStatus(m.SizeBytes, profile)
		if m.IsDefault {
			result[i].RecommendedQuant = "Q4_K_M"
		}
	}
	return result
}

func GetRecommendedModelID() string {
	return hardware.RecommendedModelID(hardware.GetProfile())
}
