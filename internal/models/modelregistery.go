package models

import docmodels "go-api/internal/models/doc-category"
import ragmodels "go-api/internal/models/rag"
import user "go-api/internal/models/user"

var MigrateModels = []interface{}{
	&user.User{},
	&user.Role{},
	&user.Permission{},
	&docmodels.DocCategory{},
	&docmodels.Document{},
	&ragmodels.Chunk{},
}
