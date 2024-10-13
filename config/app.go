package config

import "os"

var (
	AppID                   = os.Getenv("APP_ID")
	AppSecret               = os.Getenv("APP_SECRET")
	BitableAppToken         = os.Getenv("BITABLE_APP_TOKEN")
	BitableTableId          = os.Getenv("BITABLE_TABLE_ID")
	BitableViewId           = os.Getenv("BITABLE_VIEW_ID")
	DefaultItemLimitPerFeed = 5
	CardTemplateId          = os.Getenv("CARD_TEMPLATE_ID")
	CardTemplateVersionName = os.Getenv("CARD_TEMPLATE_VERSION_NAME")
	DocLink                 = "https://bqc4atlhac.feishu.cn/docx/PjPqd7Tk4o728yxqTdvc9KfanNh"
)
