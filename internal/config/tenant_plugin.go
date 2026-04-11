package config

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TenantPlugin is a GORM plugin for multi-tenancy isolation
type TenantPlugin struct{}

func (p *TenantPlugin) Name() string {
	return "TenantPlugin"
}

func (p *TenantPlugin) Initialize(db *gorm.DB) error {
	// Query
	db.Callback().Query().Before("gorm:query").Register("tenant_plugin:query", p.applyTenantFilter)
	// Update
	db.Callback().Update().Before("gorm:update").Register("tenant_plugin:update", p.applyTenantFilter)
	// Delete
	db.Callback().Delete().Before("gorm:delete").Register("tenant_plugin:delete", p.applyTenantFilter)
	// Row
	db.Callback().Row().Before("gorm:row").Register("tenant_plugin:row", p.applyTenantFilter)
	// Raw
	db.Callback().Raw().Before("gorm:raw").Register("tenant_plugin:raw", p.applyTenantFilter)
	
	return nil
}

func (p *TenantPlugin) applyTenantFilter(db *gorm.DB) {
	if db.Statement.Schema != nil {
		// Check if model has TenantID field
		if field := db.Statement.Schema.LookUpField("TenantID"); field != nil {
			// Get tenant_id from context
			tenantIDVal := db.Statement.Context.Value("tenant_id")
			if tenantIDVal != nil {
				if tenantID, ok := tenantIDVal.(uint); ok && tenantID != 0 {
					tableName := db.Statement.Table
					if tableName == "" {
						tableName = db.Statement.Schema.Table
					}

					// Apply WHERE (table.tenant_id = ? OR table.tenant_id IS NULL) 
					// Using explicit table name prefix to avoid SQL ambiguity during JOINs
					db.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
						clause.Or(
							clause.Eq{Column: clause.Column{Table: tableName, Name: "tenant_id"}, Value: tenantID},
							clause.Expr{SQL: fmt.Sprintf("%s.tenant_id IS NULL", tableName)},
						),
					}})
				}
			}
		}
	}
}
