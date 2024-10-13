package model

type FeishuMessageData struct {
	TemplateId          string                 `json:"template_id"`
	TemplateVersionName string                 `json:"template_version_name"`
	TemplateVariable    map[string]interface{} `json:"template_variable"`
}

type FeishuMessageContent struct {
	Type string            `json:"type"`
	Data FeishuMessageData `json:"data"`
}

type FeishuMessageItem struct {
	Title            string `json:"title"`
	Link             string `json:"link"`
	PrimaryDesc      string `json:"primaryDesc"`
	PrimaryDescColor string `json:"primaryDescColor"`
	SecondaryDesc    string `json:"secondaryDesc"`
}
