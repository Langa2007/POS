// bootstrap creates the first tenant and owner for a new local installation.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/platform/auth"
	"github.com/Langa2007/POS/backend/internal/platform/db"
	"github.com/jackc/pgx/v5"
)

func required(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}
func main() {
	ctx := context.Background()
	store, err := db.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	if err = store.Migrate(ctx); err != nil {
		log.Fatal(err)
	}
	tenantID, _ := ids.NewUUIDv7()
	businessID, _ := ids.NewUUIDv7()
	branchID, _ := ids.NewUUIDv7()
	userID, _ := ids.NewUUIDv7()
	roleID, _ := ids.NewUUIDv7()
	hash, err := auth.HashPassword(required("ADMIN_PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}
	err = store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO tenants(id,name) VALUES($1,$2)`, tenantID, required("TENANT_NAME")); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO businesses(id,tenant_id,name) VALUES($1,$2,$3)`, businessID, tenantID, required("BUSINESS_NAME")); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO branches(id,tenant_id,business_id,name) VALUES($1,$2,$3,$4)`, branchID, tenantID, businessID, required("BRANCH_NAME")); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO users(id,tenant_id,email,password_hash) VALUES($1,$2,lower($3),$4)`, userID, tenantID, required("ADMIN_EMAIL"), hash); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO user_branches(tenant_id,user_id,branch_id) VALUES($1,$2,$3)`, tenantID, userID, branchID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO roles(id,tenant_id,name) VALUES($1,$2,'owner')`, roleID, tenantID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO user_roles(tenant_id,user_id,role_id) VALUES($1,$2,$3)`, tenantID, userID, roleID); err != nil {
			return err
		}
		permissions := []string{"core.*", "users.manage", "settings.manage", "terminals.register"}
		for _, permission := range permissions {
			if _, err := tx.Exec(ctx, `INSERT INTO permissions(id,description) VALUES($1,'System permission') ON CONFLICT (id) DO NOTHING`, permission); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO role_permissions(tenant_id,role_id,permission_id) VALUES($1,$2,$3)`, tenantID, roleID, permission); err != nil {
				return err
			}
		}
		auditID, _ := ids.NewUUIDv7()
		_, err = tx.Exec(ctx, `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'tenant.bootstrap','tenant',$4,jsonb_build_object('business_id',$5,'branch_id',$6))`, auditID, tenantID, userID, tenantID, businessID, branchID)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tenant created: %s\nOwner user: %s\n", tenantID, userID)
}
