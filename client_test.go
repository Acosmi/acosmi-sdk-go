package acosmi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// ---------- Test Helpers ----------

// testClient creates a Client with pre-injected valid token pointing to the given test server
func testClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	c, err := NewClient(Config{ServerURL: serverURL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.tokens = &TokenSet{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		ClientID:     "test-client",
		ServerURL:    serverURL,
	}
	return c
}

// jsonHandler returns an http.Handler that responds with the given JSON body
func jsonHandler(body interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	})
}

// ============================================================================
// Default ServerURL
// ============================================================================

func TestNewClient_DefaultServerURL(t *testing.T) {
	c, err := NewClient(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.serverURL != "https://acosmi.ai" {
		t.Errorf("expected default 'https://acosmi.ai', got %q", c.serverURL)
	}
}

func TestNewClient_CustomServerURL(t *testing.T) {
	c, err := NewClient(Config{ServerURL: "http://127.0.0.1:3300"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.serverURL != "http://127.0.0.1:3300" {
		t.Errorf("expected 'http://127.0.0.1:3300', got %q", c.serverURL)
	}
}

func TestNewClient_TrailingSlashStripped(t *testing.T) {
	c, err := NewClient(Config{ServerURL: "https://acosmi.ai/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.serverURL != "https://acosmi.ai" {
		t.Errorf("expected trailing slash stripped, got %q", c.serverURL)
	}
}

// ============================================================================
// Business Error Detection (根因修复 #16)
// ============================================================================

func TestBuyTokenPackage_BusinessError(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code": 500,
		"msg":  "余额不足",
		"data": nil,
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	_, err := c.BuyTokenPackage(context.Background(), "pkg-123", nil)
	if err == nil {
		t.Fatal("expected error for business code 500, got nil")
	}

	var bizErr *BusinessError
	if !errors.As(err, &bizErr) {
		t.Fatalf("expected *BusinessError, got %T: %v", err, err)
	}
	if bizErr.Code != 500 {
		t.Errorf("expected code 500, got %d", bizErr.Code)
	}
	if bizErr.Message != "余额不足" {
		t.Errorf("expected message '余额不足', got %q", bizErr.Message)
	}
}

func TestGetBalance_BusinessError(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code":    401,
		"message": "token expired",
		"data":    nil,
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	_, err := c.GetBalance(context.Background())
	if err == nil {
		t.Fatal("expected error for business code 401, got nil")
	}

	var bizErr *BusinessError
	if !errors.As(err, &bizErr) {
		t.Fatalf("expected *BusinessError, got %T: %v", err, err)
	}
	if bizErr.Code != 401 {
		t.Errorf("expected code 401, got %d", bizErr.Code)
	}
}

func TestGetBalance_Success(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"totalTokenQuota":     1000000,
			"totalTokenUsed":      250000,
			"totalTokenRemaining": 750000,
			"totalCallQuota":      100,
			"totalCallUsed":       30,
			"totalCallRemaining":  70,
			"activeEntitlements":  3,
		},
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	bal, err := c.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bal.TotalTokenQuota != 1000000 {
		t.Errorf("expected quota 1000000, got %d", bal.TotalTokenQuota)
	}
	if bal.TotalTokenRemaining != 750000 {
		t.Errorf("expected remaining 750000, got %d", bal.TotalTokenRemaining)
	}
	if bal.ActiveEntitlements != 3 {
		t.Errorf("expected 3 active, got %d", bal.ActiveEntitlements)
	}
}

func TestBusinessError_MessageFallback(t *testing.T) {
	// 测试 yudao 格式 (msg) 和 nexus-v4 格式 (message) 的兼容
	tests := []struct {
		name     string
		resp     map[string]interface{}
		wantMsg  string
	}{
		{"nexus message field", map[string]interface{}{"code": 1, "message": "权益不足", "data": nil}, "权益不足"},
		{"yudao msg field", map[string]interface{}{"code": 1, "msg": "参数错误", "data": nil}, "参数错误"},
		{"both fields, message wins", map[string]interface{}{"code": 1, "message": "A", "msg": "B", "data": nil}, "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(jsonHandler(tt.resp))
			defer server.Close()

			c := testClient(t, server.URL)
			_, err := c.GetBalance(context.Background())

			var bizErr *BusinessError
			if !errors.As(err, &bizErr) {
				t.Fatalf("expected *BusinessError, got %T: %v", err, err)
			}
			if bizErr.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, bizErr.Message)
			}
		})
	}
}

// ============================================================================
// Path Escaping (路径注入防护)
// ============================================================================

func TestPathEscaping_BuyTokenPackage(t *testing.T) {
	var capturedURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"id": "o1", "status": "UNPAID"},
		})
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	_, _ = c.BuyTokenPackage(context.Background(), "../evil", nil)

	// url.PathEscape("../evil") = "..%2Fevil"
	if !strings.Contains(capturedURI, "..%2Fevil") {
		t.Errorf("expected path-escaped '../evil' in URI, got: %s", capturedURI)
	}
}

func TestPathEscaping_GetOrderStatus(t *testing.T) {
	var capturedURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"orderId": "x", "status": "PAID"},
		})
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	_, _ = c.GetOrderStatus(context.Background(), "../../admin")

	if !strings.Contains(capturedURI, "..%2F..%2Fadmin") {
		t.Errorf("expected path-escaped '../../admin' in URI, got: %s", capturedURI)
	}
}

func TestPathEscaping_GetTool(t *testing.T) {
	var capturedURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"id": "t1", "name": "test"},
		})
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	_, _ = c.GetTool(context.Background(), "foo/../../bar")

	if !strings.Contains(capturedURI, "foo%2F..%2F..%2Fbar") {
		t.Errorf("expected path-escaped ID in URI, got: %s", capturedURI)
	}
}

// ============================================================================
// ListTokenPackages Format Compatibility
// ============================================================================

func TestListTokenPackages_YudaoFormat(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"list": []map[string]interface{}{
				{"id": "pkg-1", "name": "基础包", "tokenQuota": 100000, "price": "9.9", "validDays": 30},
				{"id": "pkg-2", "name": "高级包", "tokenQuota": 500000, "price": "39.9", "validDays": 30},
			},
			"total": 2,
		},
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	pkgs, err := c.ListTokenPackages(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "基础包" {
		t.Errorf("expected name '基础包', got %q", pkgs[0].Name)
	}
	if pkgs[1].Name != "高级包" {
		t.Errorf("expected name '高级包', got %q", pkgs[1].Name)
	}
}

func TestListTokenPackages_ArrayFormat(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code": 0,
		"data": []map[string]interface{}{
			{"id": "pkg-1", "name": "基础包", "tokenQuota": 100000, "price": "9.9", "validDays": 30},
		},
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	pkgs, err := c.ListTokenPackages(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].ID != "pkg-1" {
		t.Errorf("expected id 'pkg-1', got %q", pkgs[0].ID)
	}
}

func TestListTokenPackages_BusinessError(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code": 500,
		"msg":  "internal error",
		"data": nil,
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	_, err := c.ListTokenPackages(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var bizErr *BusinessError
	if !errors.As(err, &bizErr) {
		t.Fatalf("expected *BusinessError, got %T: %v", err, err)
	}
}

// ============================================================================
// WaitForPayment
// ============================================================================

func TestWaitForPayment_Success(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		status := "UNPAID"
		if n >= 3 {
			status = "PAID"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"orderId": "order-123", "status": status},
		})
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := c.WaitForPayment(ctx, "order-123", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != "PAID" {
		t.Errorf("expected PAID, got %s", status.Status)
	}
	if n := atomic.LoadInt32(&callCount); n < 3 {
		t.Errorf("expected at least 3 polls, got %d", n)
	}
}

func TestWaitForPayment_TerminalFailure(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{"orderId": "order-456", "status": "CANCELLED"},
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	status, err := c.WaitForPayment(context.Background(), "order-456", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for CANCELLED order, got nil")
	}

	var termErr *OrderTerminalError
	if !errors.As(err, &termErr) {
		t.Fatalf("expected *OrderTerminalError, got %T: %v", err, err)
	}
	if termErr.Status != "CANCELLED" {
		t.Errorf("expected status CANCELLED, got %s", termErr.Status)
	}
	if status == nil {
		t.Fatal("expected non-nil status even on terminal failure")
	}
}

func TestWaitForPayment_Timeout(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{"orderId": "order-789", "status": "UNPAID"},
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := c.WaitForPayment(ctx, "order-789", 50*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestWaitForPayment_DefaultInterval(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{"orderId": "order-x", "status": "PAID"},
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	// pollInterval=0 → 默认 2s, 但首次查询立即返回 PAID
	status, err := c.WaitForPayment(context.Background(), "order-x", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != "PAID" {
		t.Errorf("expected PAID, got %s", status.Status)
	}
}

// ============================================================================
// Purchase Chain End-to-End
// ============================================================================

func TestPurchaseChain_EndToEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		switch {
		// Step 1: List packages
		case strings.HasSuffix(path, "/token-packages") && r.Method == "GET":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []map[string]interface{}{
					{"id": "pkg-basic", "name": "基础包", "tokenQuota": 100000, "price": "9.9", "validDays": 30, "isEnabled": true},
				},
			})

		// Step 2: Buy
		case strings.Contains(path, "/buy") && r.Method == "POST":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"id":          "order-e2e",
					"packageId":   "pkg-basic",
					"packageName": "基础包",
					"amount":      "9.9",
					"status":      "UNPAID",
					"payUrl":      "https://pay.example.com/order-e2e",
				},
			})

		// Step 3: Order status (immediately PAID for E2E simplicity)
		case strings.Contains(path, "/status"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{"orderId": "order-e2e", "status": "PAID"},
			})

		// Step 4: Balance
		case strings.Contains(path, "/entitlements/balance"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"totalTokenQuota":     100000,
					"totalTokenUsed":      0,
					"totalTokenRemaining": 100000,
					"activeEntitlements":  1,
				},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := testClient(t, server.URL)
	ctx := context.Background()

	// Step 1: List packages
	pkgs, err := c.ListTokenPackages(ctx)
	if err != nil {
		t.Fatalf("ListTokenPackages: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages returned")
	}
	if pkgs[0].ID != "pkg-basic" {
		t.Errorf("expected pkg-basic, got %s", pkgs[0].ID)
	}

	// Step 2: Buy
	order, err := c.BuyTokenPackage(ctx, pkgs[0].ID, nil)
	if err != nil {
		t.Fatalf("BuyTokenPackage: %v", err)
	}
	if order.ID != "order-e2e" {
		t.Errorf("expected order-e2e, got %s", order.ID)
	}
	if order.PayURL == "" {
		t.Error("expected payUrl")
	}

	// Step 3: Wait for payment
	status, err := c.WaitForPayment(ctx, order.ID, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForPayment: %v", err)
	}
	if status.Status != "PAID" {
		t.Errorf("expected PAID, got %s", status.Status)
	}

	// Step 4: Verify balance
	bal, err := c.GetBalance(ctx)
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if bal.TotalTokenRemaining != 100000 {
		t.Errorf("expected 100000 tokens, got %d", bal.TotalTokenRemaining)
	}
	if bal.ActiveEntitlements != 1 {
		t.Errorf("expected 1 active entitlement, got %d", bal.ActiveEntitlements)
	}
}
