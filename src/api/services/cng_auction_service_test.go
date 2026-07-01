package services

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCNGAuctionService_ParseLotPage(t *testing.T) {
	svc := NewCNGAuctionService(nil)
	lot, err := svc.parseLotPage(cngLotFixture())
	if err != nil {
		t.Fatalf("parseLotPage returned error: %v", err)
	}

	if lot.SourceLotID != "4-LOTID" {
		t.Fatalf("SourceLotID = %q, want 4-LOTID", lot.SourceLotID)
	}
	if lot.SourceSaleID != "4-SALEID" {
		t.Fatalf("SourceSaleID = %q, want 4-SALEID", lot.SourceSaleID)
	}
	if lot.LotNumber != 43 {
		t.Fatalf("LotNumber = %d, want 43", lot.LotNumber)
	}
	if lot.URL != "https://auctions.cngcoins.com/lots/view/4-LOTID/test-lot" {
		t.Fatalf("URL = %q", lot.URL)
	}
	if lot.ImageURL != "https://images.example/43_1.jpg" {
		t.Fatalf("ImageURL = %q", lot.ImageURL)
	}
	if lot.Estimate == nil || *lot.Estimate != 100 {
		t.Fatalf("Estimate = %v, want 100", lot.Estimate)
	}
	if lot.CurrentBid == nil || *lot.CurrentBid != 60 {
		t.Fatalf("CurrentBid = %v, want 60", lot.CurrentBid)
	}
	if lot.Currency != "USD" {
		t.Fatalf("Currency = %q, want USD", lot.Currency)
	}
	if lot.SaleName != "Electronic Auction 612" {
		t.Fatalf("SaleName = %q", lot.SaleName)
	}
	if lot.Description == "" || strings.Contains(lot.Description, "<b>") {
		t.Fatalf("Description was not cleaned: %q", lot.Description)
	}
}

func TestCNGAuctionService_ParseWatchlist(t *testing.T) {
	svc := NewCNGAuctionService(nil)
	lots := svc.ParseWatchlist(cngWatchlistFixture())

	if len(lots) != 2 {
		t.Fatalf("ParseWatchlist returned %d lots, want 2", len(lots))
	}
	if lots[0].SourceLotID != "4-LOT1" || lots[1].SourceLotID != "4-LOT2" {
		t.Fatalf("unexpected source lot IDs: %#v", lots)
	}
	if lots[1].Currency != "USD" {
		t.Fatalf("second lot currency = %q, want USD fallback", lots[1].Currency)
	}
}

func TestCNGAuctionService_LoginAndFetchWatchlist(t *testing.T) {
	var loggedIn bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			if !loggedIn {
				w.Write([]byte(`viewVars = {"me":null};`))
				return
			}
			w.Write([]byte(`viewVars = {"me":{"row_id":"user"}};`))
			return
		case "/login":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<form action="/login"><input name="username"><input name="password"></form>`))
				return
			}
			if r.Method == http.MethodPost {
				if err := r.ParseForm(); err != nil {
					t.Fatalf("ParseForm failed: %v", err)
				}
				if r.Form.Get("username") != "user@example.com" || r.Form.Get("password") != "secret" || r.Form.Get("Login") != "Login" {
					t.Fatalf("unexpected login form: %#v", r.Form)
				}
				loggedIn = true
				http.SetCookie(w, &http.Cookie{Name: "PHPSESSID", Value: "test"})
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		case "/ajax/refresh-me":
			if !loggedIn {
				w.Write([]byte(`null`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"row_id":"user"}`))
			return
		case "/watched-lots":
			if !loggedIn {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			w.Write([]byte(cngWatchlistFixture()))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	restore := overrideCNGURLs(server.URL)
	defer restore()

	svc := NewCNGAuctionService(nil)
	client, err := svc.Login("user@example.com", "secret")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	raw, err := svc.FetchWatchlist(client)
	if err != nil {
		t.Fatalf("FetchWatchlist returned error: %v", err)
	}
	if got := svc.ParseWatchlist(raw); len(got) != 2 {
		t.Fatalf("parsed %d watched lots, want 2", len(got))
	}
}

func TestCNGAuctionService_LoginInvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<form action="/login"><input name="username"><input name="password"></form>`))
		case "/ajax/refresh-me":
			w.Write([]byte(`null`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	restore := overrideCNGURLs(server.URL)
	defer restore()

	svc := NewCNGAuctionService(nil)
	if _, err := svc.Login("bad@example.com", "wrong"); err == nil {
		t.Fatal("Login succeeded, want error")
	}
}

func overrideCNGURLs(base string) func() {
	oldLoginURL := cngLoginURL
	oldWatchlistURL := cngWatchlistURL
	oldRefreshMeURL := cngRefreshMeURL
	cngLoginURL = base + "/login"
	cngWatchlistURL = base + "/watched-lots"
	cngRefreshMeURL = base + "/ajax/refresh-me"
	return func() {
		cngLoginURL = oldLoginURL
		cngWatchlistURL = oldWatchlistURL
		cngRefreshMeURL = oldRefreshMeURL
	}
}

func cngLotFixture() string {
	return `<!doctype html><html><script>
viewVars = {
  "currentRouteName":"lot-detail-slug",
  "lot":{
    "row_id":"4-LOTID",
    "lot_number":43,
    "lot_number_extension":"",
    "title":"CARTHAGE. Half-Shekel. Good VF.",
    "description":"<b>CARTHAGE.</b> Second Punic War. Good VF.",
    "estimate_low":"100.00",
    "estimate_high":"150.00",
    "currency_code":"USD",
    "starting_price":"60.00",
    "sold_price":null,
    "status":"active",
    "_detail_url":"/lots/view/4-LOTID/test-lot",
    "cover_thumbnail":"",
    "images":[{"detail_url":"https://images.example/43_1.jpg","thumbnail_url":"https://images.example/thumb.jpg"}],
    "auction":{
      "row_id":"4-SALEID",
      "title":"Electronic Auction 612",
      "currency_code":"USD",
      "time_start":"2026-06-17T20:00:00Z",
      "effective_end_time":"2026-07-01T21:15:00Z"
    }
  }
};
</script></html>`
}

func cngWatchlistFixture() string {
	return `<!doctype html><html><script>
viewVars = {
  "currentRouteName":"watched-lots-index",
  "lots":{
    "query_info":{"total_num_results":2,"page_size":48},
    "result_page":[
      {
        "row_id":"4-LOT1",
        "lot_number":1,
        "title":"Lot One",
        "truncated_description":"<b>Lot</b> one description",
        "estimate_low":"100.00",
        "currency_code":"USD",
        "starting_price":"60.00",
        "_detail_url":"/lots/view/4-LOT1/lot-one",
        "cover_thumbnail":"https://images.example/1.jpg",
        "auction":{"row_id":"4-SALEID","title":"Electronic Auction 612","currency_code":"USD","effective_end_time":"2026-07-01T21:15:00Z"}
      },
      {
        "row_id":"4-LOT2",
        "lot_number":2,
        "title":"Lot Two",
        "estimate_low":"200.00",
        "starting_price":"120.00",
        "_detail_url":"/lots/view/4-LOT2/lot-two",
        "cover_thumbnail":"https://images.example/2.jpg",
        "auction":{"row_id":"4-SALEID","title":"Electronic Auction 612","currency_code":"USD","effective_end_time":"2026-07-01T21:15:00Z"}
      }
    ]
  }
};
</script></html>`
}
