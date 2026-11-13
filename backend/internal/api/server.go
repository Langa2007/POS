package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Langa2007/POS/backend/internal/adapters/mpesa"
	"github.com/Langa2007/POS/backend/internal/core/customfields"
	"github.com/Langa2007/POS/backend/internal/core/fiscal"
	"github.com/Langa2007/POS/backend/internal/core/importexport"
	"github.com/Langa2007/POS/backend/internal/core/presets"
	"github.com/Langa2007/POS/backend/internal/core/rules"
	"github.com/Langa2007/POS/backend/internal/core/templates"
	"github.com/Langa2007/POS/backend/internal/core/workflows"
	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/Langa2007/POS/backend/internal/platform/auth"
	"github.com/Langa2007/POS/backend/internal/platform/db"
	"github.com/jackc/pgx/v5"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type Server struct {
	db              *db.Store
	tokens          *auth.TokenManager
	logger          *slog.Logger
	credentialKey   []byte
	mpesaMu         sync.Mutex
	mpesaProviders  map[string]cachedMPesaProvider
	fiscalSubmitter fiscal.Submitter
	presets         *presets.Store
	customfields    *customfields.Store
	rules           *rules.Store
	workflows       *workflows.Store
	templates       *templates.Store
	importexport    *importexport.Store
}

type cachedMPesaProvider struct {
	fingerprint [32]byte
	provider    *mpesa.Daraja
	config      mpesaStoredConfig
}

func New(store *db.Store, tokens *auth.TokenManager, logger *slog.Logger) *Server {
	key := []byte(strings.TrimSpace(os.Getenv("POS_CREDENTIALS_KEY")))
	if decoded, err := base64.StdEncoding.DecodeString(string(key)); err == nil && len(decoded) == 32 {
		key = decoded
	}
	s := &Server{
		db:             store,
		tokens:         tokens,
		logger:         logger,
		credentialKey:  key,
		mpesaProviders: make(map[string]cachedMPesaProvider),
		presets:        presets.New(store.Pool()),
		customfields:   customfields.New(store.Pool()),
		rules:          rules.New(store.Pool()),
		workflows:      workflows.New(store.Pool()),
		templates:      templates.New(store.Pool()),
		importexport:   importexport.New(store.Pool()),
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("POS_FISCAL_MODE")), "simulator") {
		s.fiscalSubmitter = fiscal.Simulator{}
	}
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", s.live)
	mux.HandleFunc("GET /health/ready", s.ready)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refresh)
	mux.HandleFunc("POST /api/v1/auth/pin-login", s.pinLogin)
	terminalRegistration := s.requirePermission("terminals.register", http.HandlerFunc(s.registerTerminal))
	mux.Handle("POST /api/v1/terminals/register", s.requireAuth(terminalRegistration))
	setPIN := s.requirePermission("users.manage", http.HandlerFunc(s.setUserPIN))
	mux.Handle("PUT /api/v1/users/{userID}/pin", s.requireAuth(setPIN))
	manifest := s.requirePermission("core.manifest", http.HandlerFunc(s.manifest))
	mux.Handle("GET /api/v1/me/manifest", s.requireAuth(manifest))
	mux.Handle("GET /api/v1/catalog/items", s.requireAuth(s.requirePermission("catalog.read", http.HandlerFunc(s.catalog))))
	mux.Handle("GET /api/v1/pharmacy/catalog", s.requireAuth(s.requirePermission("pharmacy.read", http.HandlerFunc(s.pharmacyCatalog))))
	mux.Handle("GET /api/v1/pharmacy/prescriptions", s.requireAuth(s.requirePermission("pharmacy.read", http.HandlerFunc(s.emptyPharmacyResource))))
	mux.Handle("GET /api/v1/pharmacy/batches", s.requireAuth(s.requirePermission("pharmacy.read", http.HandlerFunc(s.emptyPharmacyResource))))
	mux.Handle("GET /api/v1/pharmacy/controlled", s.requireAuth(s.requirePermission("pharmacy.read", http.HandlerFunc(s.emptyPharmacyResource))))
	mux.Handle("GET /api/v1/pharmacy/formulary", s.requireAuth(s.requirePermission("pharmacy.read", http.HandlerFunc(s.emptyPharmacyResource))))
	mux.Handle("GET /api/v1/pharmacy/labels", s.requireAuth(s.requirePermission("pharmacy.read", http.HandlerFunc(s.emptyPharmacyResource))))
	mux.Handle("GET /api/v1/restaurant/menu", s.requireAuth(s.requirePermission("restaurant.read", http.HandlerFunc(s.restaurantMenu))))
	mux.Handle("GET /api/v1/restaurant/kitchen/tickets", s.requireAuth(s.requirePermission("restaurant.read", http.HandlerFunc(s.restaurantKitchenTickets))))
	mux.Handle("GET /api/v1/restaurant/tables", s.requireAuth(s.requirePermission("restaurant.read", http.HandlerFunc(s.restaurantTables))))
	mux.Handle("PATCH /api/v1/restaurant/tables/{tableID}", s.requireAuth(s.requirePermission("restaurant.order", http.HandlerFunc(s.updateRestaurantTable))))
	mux.Handle("GET /api/v1/restaurant/summary", s.requireAuth(s.requirePermission("restaurant.read", http.HandlerFunc(s.restaurantSummary))))
	mux.Handle("POST /api/v1/catalog/items", s.requireAuth(s.requirePermission("catalog.manage", http.HandlerFunc(s.createItem))))
	mux.Handle("PUT /api/v1/catalog/items/{itemID}", s.requireAuth(s.requirePermission("catalog.manage", http.HandlerFunc(s.updateItem))))
	mux.Handle("DELETE /api/v1/catalog/items/{itemID}", s.requireAuth(s.requirePermission("catalog.manage", http.HandlerFunc(s.deleteItem))))
	mux.Handle("POST /api/v1/catalog/items/{itemID}/barcodes", s.requireAuth(s.requirePermission("catalog.manage", http.HandlerFunc(s.addItemBarcode))))
	mux.Handle("POST /api/v1/shifts", s.requireAuth(s.requirePermission("sales.shift", http.HandlerFunc(s.openShift))))
	mux.Handle("POST /api/v1/shifts/{shiftID}/close", s.requireAuth(s.requirePermission("sales.shift", http.HandlerFunc(s.closeShift))))
	mux.Handle("POST /api/v1/sales", s.requireAuth(s.requirePermission("sales.create", http.HandlerFunc(s.createSale))))
	mux.Handle("GET /api/v1/reports/daily", s.requireAuth(s.requirePermission("reports.read", http.HandlerFunc(s.dailyReport))))
	mux.Handle("POST /api/v1/sales/{saleID}/payments", s.requireAuth(s.requirePermission("payments.create", http.HandlerFunc(s.allocatePayment))))
	mux.Handle("POST /api/v1/payments/intents", s.requireAuth(s.requirePermission("payments.create", http.HandlerFunc(s.initiatePayment))))
	mux.Handle("GET /api/v1/payments/intents/{intentID}", s.requireAuth(s.requirePermission("payments.read", http.HandlerFunc(s.paymentIntentStatus))))
	mux.Handle("PUT /api/v1/payments/providers/mpesa", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.configureMPesa))))
	mux.Handle("GET /api/v1/payments/providers/mpesa", s.requireAuth(s.requirePermission("payments.read", http.HandlerFunc(s.getMPesaConfig))))
	mux.Handle("POST /api/v1/payments/providers/mpesa/c2b/register", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.registerMPesaC2B))))
	mux.Handle("POST /api/v1/payments/refunds/{saleID}", s.requireAuth(s.requirePermission("payments.refund", http.HandlerFunc(s.refundPayment))))
	mux.Handle("POST /api/v1/payments/refunds/{refundID}/approve", s.requireAuth(s.requirePermission("payments.refund.approve", http.HandlerFunc(s.approveRefund))))
	mux.Handle("GET /api/v1/payments", s.requireAuth(s.requirePermission("payments.read", http.HandlerFunc(s.paymentList))))
	mux.Handle("GET /api/v1/ledger/trial-balance", s.requireAuth(s.requirePermission("ledger.read", http.HandlerFunc(s.trialBalance))))
	mux.Handle("POST /api/v1/reconciliation/import", s.requireAuth(s.requirePermission("payments.reconcile", http.HandlerFunc(s.reconcile))))
	mux.Handle("GET /api/v1/reconciliation", s.requireAuth(s.requirePermission("payments.reconcile", http.HandlerFunc(s.reconciliationList))))
	mux.Handle("GET /api/v1/payments/unallocated", s.requireAuth(s.requirePermission("payments.read", http.HandlerFunc(s.unallocatedPaymentList))))
	mux.Handle("POST /api/v1/payments/unallocated/{paymentID}/allocate", s.requireAuth(s.requirePermission("payments.create", http.HandlerFunc(s.allocateUnallocatedPayment))))
	mux.Handle("POST /api/v1/credit/accounts", s.requireAuth(s.requirePermission("payments.create", http.HandlerFunc(s.createCreditAccount))))
	mux.Handle("GET /api/v1/credit/accounts/{accountID}", s.requireAuth(s.requirePermission("payments.read", http.HandlerFunc(s.creditStatement))))
	mux.Handle("POST /api/v1/sync/push", s.requireAuth(s.requirePermission("sync.push", http.HandlerFunc(s.syncPush))))
	mux.Handle("GET /api/v1/sync/pull", s.requireAuth(s.requirePermission("sync.pull", http.HandlerFunc(s.syncPull))))
	mux.Handle("GET /api/v1/sync/policy", s.requireAuth(s.requirePermission("sync.policy.read", http.HandlerFunc(s.getSyncPolicy))))
	mux.Handle("PUT /api/v1/sync/policy", s.requireAuth(s.requirePermission("sync.policy.manage", http.HandlerFunc(s.updateSyncPolicy))))
	mux.Handle("GET /api/v1/tax/rates", s.requireAuth(s.requirePermission("tax.read", http.HandlerFunc(s.taxRates))))
	mux.Handle("POST /api/v1/tax/rates", s.requireAuth(s.requirePermission("tax.manage", http.HandlerFunc(s.putTaxRate))))
	mux.Handle("POST /api/v1/tax/calculate", s.requireAuth(s.requirePermission("tax.read", http.HandlerFunc(s.calculateTax))))
	mux.Handle("GET /api/v1/suppliers", s.requireAuth(s.requirePermission("purchasing.read", http.HandlerFunc(s.suppliers))))
	mux.Handle("POST /api/v1/suppliers", s.requireAuth(s.requirePermission("purchasing.manage", http.HandlerFunc(s.suppliers))))
	mux.Handle("GET /api/v1/purchase-orders", s.requireAuth(s.requirePermission("purchasing.read", http.HandlerFunc(s.purchaseOrders))))
	mux.Handle("POST /api/v1/purchase-orders", s.requireAuth(s.requirePermission("purchasing.manage", http.HandlerFunc(s.purchaseOrders))))
	mux.Handle("POST /api/v1/purchase-orders/{orderID}/approve", s.requireAuth(s.requirePermission("purchasing.approve", http.HandlerFunc(s.approvePurchaseOrder))))
	mux.Handle("POST /api/v1/purchase-orders/{orderID}/receipts", s.requireAuth(s.requirePermission("purchasing.receive", http.HandlerFunc(s.receivePurchaseOrder))))
	mux.Handle("GET /api/v1/purchasing/reorder-suggestions", s.requireAuth(s.requirePermission("purchasing.read", http.HandlerFunc(s.reorderSuggestions))))
	mux.Handle("GET /api/v1/reports/supply-summary", s.requireAuth(s.requirePermission("reports.read", http.HandlerFunc(s.supplySummary))))
	mux.Handle("POST /api/v1/stock-takes", s.requireAuth(s.requirePermission("inventory.count", http.HandlerFunc(s.createStockTake))))
	mux.Handle("POST /api/v1/stock-takes/{takeID}/approve", s.requireAuth(s.requirePermission("inventory.adjust", http.HandlerFunc(s.approveStockTake))))
	mux.Handle("GET /api/v1/stock-transfers", s.requireAuth(s.requirePermission("inventory.read", http.HandlerFunc(s.stockTransfers))))
	mux.Handle("POST /api/v1/stock-transfers", s.requireAuth(s.requirePermission("inventory.transfer", http.HandlerFunc(s.stockTransfers))))
	mux.Handle("POST /api/v1/stock-transfers/{transferID}/receive", s.requireAuth(s.requirePermission("inventory.transfer", http.HandlerFunc(s.receiveStockTransfer))))
	mux.Handle("POST /api/v1/supplier-invoices", s.requireAuth(s.requirePermission("payables.manage", http.HandlerFunc(s.createSupplierInvoice))))
	mux.Handle("GET /api/v1/supplier-invoices", s.requireAuth(s.requirePermission("payables.read", http.HandlerFunc(s.supplierInvoices))))
	mux.Handle("POST /api/v1/supplier-invoices/{invoiceID}/payments", s.requireAuth(s.requirePermission("payables.pay", http.HandlerFunc(s.paySupplierInvoice))))
	mux.Handle("POST /api/v1/supplier-returns", s.requireAuth(s.requirePermission("purchasing.return", http.HandlerFunc(s.createSupplierReturn))))
	mux.Handle("POST /api/v1/fiscal/submissions", s.requireAuth(s.requirePermission("fiscal.submit", http.HandlerFunc(s.createFiscalSubmission))))
	mux.Handle("POST /api/v1/fiscal/submissions/{submissionID}/retry", s.requireAuth(s.requirePermission("fiscal.submit", http.HandlerFunc(s.retryFiscalSubmission))))
	mux.HandleFunc("POST /api/v1/callbacks/mpesa/stk/{token}", s.mpesaCallback)
	mux.HandleFunc("POST /api/v1/callbacks/mpesa/c2b/{token}", s.mpesaC2BCallback)
	mux.HandleFunc("POST /api/v1/callbacks/mpesa/c2b/validate/{token}", s.mpesaC2BValidation)
	mux.Handle("GET /api/v1/presets", s.requireAuth(http.HandlerFunc(s.listPresets)))
	mux.Handle("POST /api/v1/presets/{presetID}/apply", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.applyPreset))))
	mux.Handle("GET /api/v1/custom-fields", s.requireAuth(http.HandlerFunc(s.listCustomFields)))
	mux.Handle("POST /api/v1/custom-fields", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.upsertCustomField))))
	mux.Handle("GET /api/v1/rules", s.requireAuth(http.HandlerFunc(s.listRules)))
	mux.Handle("POST /api/v1/rules", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.upsertRule))))
	mux.Handle("POST /api/v1/rules/simulate", s.requireAuth(http.HandlerFunc(s.simulateRules)))
	mux.Handle("GET /api/v1/workflows", s.requireAuth(http.HandlerFunc(s.listWorkflows)))
	mux.Handle("POST /api/v1/workflows", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.upsertWorkflow))))
	mux.Handle("GET /api/v1/templates", s.requireAuth(http.HandlerFunc(s.listTemplates)))
	mux.Handle("POST /api/v1/templates", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.upsertTemplate))))
	mux.Handle("POST /api/v1/templates/{key}/render", s.requireAuth(http.HandlerFunc(s.renderTemplate)))
	mux.Handle("POST /api/v1/import/jobs", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.createImportJob))))
	mux.Handle("POST /api/v1/import/jobs/{jobID}/dry-run", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.dryRunImport))))
	mux.Handle("GET /api/v1/capabilities", s.requireAuth(http.HandlerFunc(s.listCapabilities)))
	mux.Handle("GET /api/v1/reports/consolidated", s.requireAuth(s.requirePermission("reports.read", http.HandlerFunc(s.consolidatedReport))))
	mux.Handle("POST /api/v1/payouts", s.requireAuth(s.requirePermission("payouts.create", http.HandlerFunc(s.createPayout))))
	mux.Handle("POST /api/v1/payouts/{payoutID}/approve", s.requireAuth(s.requirePermission("payouts.approve", http.HandlerFunc(s.approvePayout))))
	mux.Handle("GET /api/v1/verticals", s.requireAuth(http.HandlerFunc(s.listVerticals)))
	mux.Handle("PUT /api/v1/verticals/{verticalID}", s.requireAuth(s.requirePermission("settings.manage", http.HandlerFunc(s.setVerticalEnabled))))
	mux.Handle("GET /api/v1/salon/appointments", s.requireAuth(s.requirePermission("salon.read", http.HandlerFunc(s.listSalonAppointments))))
	mux.Handle("PATCH /api/v1/salon/appointments/{appointmentID}", s.requireAuth(s.requirePermission("salon.manage", http.HandlerFunc(s.updateSalonAppointmentStatus))))
	mux.Handle("GET /api/v1/salon/loyalty", s.requireAuth(s.requirePermission("salon.read", http.HandlerFunc(s.listSalonLoyaltyMembers))))
	mux.Handle("POST /api/v1/salon/loyalty", s.requireAuth(s.requirePermission("salon.manage", http.HandlerFunc(s.createSalonLoyaltyMember))))
	mux.Handle("POST /api/v1/salon/loyalty/{memberID}/redeem", s.requireAuth(s.requirePermission("salon.manage", http.HandlerFunc(s.redeemSalonLoyalty))))
	mux.Handle("GET /api/v1/salon/packages", s.requireAuth(s.requirePermission("salon.read", http.HandlerFunc(s.listSalonPackages))))
	mux.Handle("GET /api/v1/salon/staff", s.requireAuth(s.requirePermission("salon.read", http.HandlerFunc(s.listSalonStaff))))
	mux.Handle("GET /api/v1/salon/dashboard", s.requireAuth(s.requirePermission("salon.read", http.HandlerFunc(s.salonDashboard))))
	mux.Handle("GET /api/v1/salon/commissions", s.requireAuth(s.requirePermission("salon.read", http.HandlerFunc(s.listSalonCommissions))))
	mux.Handle("GET /api/v1/demo/hello", s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"message": "hello from demo vertical"})
	})))
	return requestID(recoverer(s.logger, mux))
}

func (s *Server) live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.db.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database.unavailable", "Database is unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type loginRequest struct {
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	var input loginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || !uuidPattern.MatchString(input.TenantID) || strings.TrimSpace(input.Email) == "" || len(input.Password) < 1 {
		writeError(w, http.StatusBadRequest, "request.invalid", "Provide tenant_id, email, and password")
		return
	}
	var userID, passwordHash string
	refreshToken, err := auth.NewRefreshToken(input.TenantID)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	refreshID, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	err = s.db.TenantTx(r.Context(), input.TenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(), `SELECT id::text,password_hash FROM users WHERE tenant_id=$1 AND lower(email)=lower($2) AND active`, input.TenantID, strings.TrimSpace(input.Email)).Scan(&userID, &passwordHash); err != nil {
			return err
		}
		if !auth.VerifyPassword(passwordHash, input.Password) {
			return errors.New("invalid credentials")
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO refresh_tokens(id,tenant_id,user_id,token_hash,expires_at) VALUES($1,$2,$3,$4,now()+interval '30 days')`, refreshID, input.TenantID, userID, auth.HashRefreshToken(refreshToken))
		return err
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.invalid_credentials", "Email or password is incorrect")
		return
	}
	token, err := s.tokens.Issue(input.TenantID, userID)
	if err != nil {
		s.logger.Error("issue access token", "error", err)
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"access_token": token, "refresh_token": refreshToken, "token_type": "Bearer", "expires_in": 900, "user_id": userID, "tenant_id": input.TenantID})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024)
	var input refreshRequest
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "request.invalid", "A refresh token is required")
		return
	}
	parts := strings.SplitN(input.RefreshToken, ".", 2)
	if len(parts) != 2 || !uuidPattern.MatchString(parts[0]) || len(parts[1]) < 40 {
		writeError(w, 400, "request.invalid", "A refresh token is required")
		return
	}
	tenantID := parts[0]
	newRefresh, err := auth.NewRefreshToken(tenantID)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	newID, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	oldHash := auth.HashRefreshToken(input.RefreshToken)
	var userID string
	var branchID, terminalID sql.NullString
	err = s.db.TenantTx(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(), `SELECT user_id::text,branch_id::text,terminal_id::text FROM refresh_tokens rt WHERE tenant_id=$1 AND token_hash=$2 AND revoked_at IS NULL AND expires_at>now() AND (terminal_id IS NULL OR EXISTS(SELECT 1 FROM terminals t WHERE t.tenant_id=rt.tenant_id AND t.id=rt.terminal_id AND t.revoked_at IS NULL)) FOR UPDATE`, tenantID, oldHash).Scan(&userID, &branchID, &terminalID); err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(), `UPDATE refresh_tokens SET revoked_at=now() WHERE tenant_id=$1 AND token_hash=$2`, tenantID, oldHash); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO refresh_tokens(id,tenant_id,user_id,branch_id,terminal_id,token_hash,expires_at) VALUES($1,$2,$3,$4,$5,$6,now()+interval '30 days')`, newID, tenantID, userID, branchID, terminalID, auth.HashRefreshToken(newRefresh))
		return err
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.invalid_refresh_token", "Refresh token is invalid or expired")
		return
	}
	access, err := s.tokens.IssueForScope(tenantID, userID, branchID.String, terminalID.String)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"access_token": access, "refresh_token": newRefresh, "token_type": "Bearer", "expires_in": 900})
}

type claimsKey struct{}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.Bearer(r.Header.Get("Authorization"))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
			return
		}
		claims, err := s.tokens.Verify(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "auth.invalid_token", "A valid bearer token is required")
			return
		}
		scope := tenancy.Scope{TenantID: claims.TenantID, UserID: claims.UserID, BranchID: claims.BranchID, TerminalID: claims.TerminalID}
		r = r.WithContext(tenancy.WithScope(context.WithValue(r.Context(), claimsKey{}, claims), scope))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) manifest(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	modules := make([]string, 0)
	capabilities := make([]string, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT module_id FROM tenant_modules WHERE tenant_id=$1 AND enabled ORDER BY module_id`, scope.TenantID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			modules = append(modules, id)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		rows, err = tx.Query(r.Context(), `SELECT capability_id FROM tenant_capabilities WHERE tenant_id=$1 AND enabled ORDER BY capability_id`, scope.TenantID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			capabilities = append(capabilities, id)
		}
		err = rows.Err()
		rows.Close()
		return err
	})
	if err != nil {
		s.logger.Error("load tenant manifest", "tenant_id", scope.TenantID, "error", err)
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tenant_id": scope.TenantID, "user_id": scope.UserID, "modules": modules, "capabilities": capabilities, "navigation": []any{}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = time.Now().UTC().Format("20060102T150405.000000000")
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}
func recoverer(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("http handler panic", "panic", v)
				writeError(w, http.StatusInternalServerError, "internal.error", "Request could not be completed")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
