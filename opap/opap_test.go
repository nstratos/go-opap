package opap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	// client is the Opap client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server

	// mux is the HTTP request multiplexer that the test HTTP server will use
	// to mock API responses.
	mux *http.ServeMux
)

func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// Opap client configured to use test server
	client = NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL)
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil)

	// test default base URL
	if got, want := c.BaseURL.String(), defaultBaseURL; got != want {
		t.Errorf("NewClient.BaseURL = %v, want %v", got, want)
	}

	// test draws default endpoint
	if got, want := c.Draws.Endpoint, defaultDrawsEndpoint; got != want {
		t.Errorf("NewClient.Draws.Endpoint = %v, want %v", got, want)
	}
}

func TestClient_NewRequest(t *testing.T) {
	c := NewClient(nil)

	inURL, outURL := "/foo", defaultBaseURL+"foo"

	req, _ := c.NewRequest("GET", inURL, nil)

	// test that the endpoint URL was correctly added to the base URL
	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("NewRequest(%q) URL =  %v, want %v", inURL, got, want)
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method = %v, want %v", got, want)
	}
}

func TestClient_Do(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		Bar string `xml:"bar"`
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"bar":"baz"}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)

	f := new(foo)
	_, err := client.Do(req, f)
	if err != nil {
		t.Fatalf("Do() returned err = %v", err)
	}

	want := &foo{"baz"}
	if got := f; !reflect.DeepEqual(got, want) {
		t.Errorf("Do() response body = %v, want %v", got, want)
	}
}

func TestClient_Do_notFound(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	req, _ := client.NewRequest("GET", "/", nil)

	resp, err := client.Do(req, nil)

	if err == nil {
		t.Error("Expected HTTP 404 error.")
	}

	if resp == nil {
		t.Fatal("Expected HTTP 404 error to return response.")
	}
}

func TestClient_Do_connectionRefused(t *testing.T) {
	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(req, nil)
	if err == nil {
		t.Error("Expected connection refused error.")
	}
}

func TestClient_NewRequest_badEndpoint(t *testing.T) {
	c := NewClient(nil)
	inURL := "%foo"
	_, err := c.NewRequest("GET", inURL, nil)
	if err == nil {
		t.Errorf("NewRequest(%q) should return parse err", inURL)
	}
}

func TestClient_Do_unexpectedHTML(t *testing.T) {
	setup()
	defer teardown()

	type foo struct{}

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `<html>something broke</html>`)
	})

	req, _ := client.NewRequest("GET", "/", nil)

	f := new(foo)
	_, err := client.Do(req, f)
	if err == nil {
		t.Fatal("Do() with unexpected HTML expected to return JSON decode err.")
	}
}

func TestDrawService_Latest(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/joker/last.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"draw":{"drawTime":"24-12-2017T22:00:00","drawNo":1873,"results":[40,13,1,24,15,8]}}`)
	})

	var game Game = Joker
	d, _, err := client.Draws.Latest(game)
	if err != nil {
		t.Fatal("client.Draws.Latest returned err:", err)
	}
	want := &Draw{DrawTime: "24-12-2017T22:00:00", DrawNo: 1873, Results: []int{40, 13, 1, 24, 15, 8}}
	if got := d; !reflect.DeepEqual(got, want) {
		t.Errorf("client.Draws.Latest(%q) \nhave: %#v\nwant: %#v", game, got, want)
	}
}

func TestDrawService_Latest_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/joker/last.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something broke", 500)
	})

	var game Game = Joker
	_, resp, err := client.Draws.Latest(game)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := resp.StatusCode, 500; got != want {
		t.Errorf("resp status code = %d, want %d", got, want)
	}
}

func TestDrawService_Latest_emptyObject(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/joker/last.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{}`)
	})

	var game Game = Joker
	d, _, err := client.Draws.Latest(game)
	if err != nil {
		t.Fatal("client.Draws.Latest returned err:", err)
	}
	want := &Draw{}
	if got := d; !reflect.DeepEqual(got, want) {
		t.Errorf("client.Draws.Latest(%q) \nhave: %#v\nwant: %#v", game, got, want)
	}
}

func TestDrawService_ByNumber(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/joker/1873.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"draw":{"drawTime":"24-12-2017T22:00:00","drawNo":1873,"results":[40,13,1,24,15,8]}}`)
	})

	var game Game = Joker
	var number = 1873
	d, _, err := client.Draws.ByNumber(game, number)
	if err != nil {
		t.Fatal("client.Draws.ByNumber returned err:", err)
	}
	want := &Draw{DrawTime: "24-12-2017T22:00:00", DrawNo: 1873, Results: []int{40, 13, 1, 24, 15, 8}}
	if got := d; !reflect.DeepEqual(got, want) {
		t.Errorf("client.Draws.ByNumber(%q, %d) \nhave: %#v\nwant: %#v", game, number, got, want)
	}
}

func TestDrawService_ByNumber_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/joker/1873.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something broke", 500)
	})

	var game Game = Joker
	_, resp, err := client.Draws.ByNumber(game, 1873)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := resp.StatusCode, 500; got != want {
		t.Errorf("resp status code = %d, want %d", got, want)
	}
}

func TestDrawService_ByDate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/joker/drawDate/24-12-2017.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"draws":{"draw":[{"drawTime":"24-12-2017T22:00:00","drawNo":1873,"results":[40,13,1,24,15,8]}]}}`)
	})

	var game Game = Joker
	day, month, year := 24, 12, 2017
	d, _, err := client.Draws.ByDate(game, day, month, year)
	if err != nil {
		t.Fatal("client.Draws.ByDate returned err:", err)
	}
	want := []Draw{{DrawTime: "24-12-2017T22:00:00", DrawNo: 1873, Results: []int{40, 13, 1, 24, 15, 8}}}
	if got := d; !reflect.DeepEqual(got, want) {
		t.Errorf("client.Draws.ByDate(%q, %d, %d, %d) \nhave: %#v\nwant: %#v", game, day, month, year, got, want)
	}
}

func TestDrawService_ByDate_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/joker/drawDate/24-12-2017.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something broke", 500)
	})

	var game Game = Joker
	_, resp, err := client.Draws.ByDate(game, 24, 12, 2017)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := resp.StatusCode, 500; got != want {
		t.Errorf("resp status code = %d, want %d", got, want)
	}
}

func TestDrawService_PropoLatest(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/proposat/last.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"draw":{"drawTime":"23-12-2017T16:00:00","drawNo":201751,"results":["2","2","1","X","X","1","X","2","1","1","1","X","2","2"]}}`)
	})

	var game PropoGame = PropoSat
	d, _, err := client.Draws.PropoLatest(game)
	if err != nil {
		t.Fatal("client.Draws.PropoLatest returned err:", err)
	}
	want := &PropoDraw{DrawTime: "23-12-2017T16:00:00", DrawNo: 201751, Results: []string{"2", "2", "1", "X", "X", "1", "X", "2", "1", "1", "1", "X", "2", "2"}}
	if got := d; !reflect.DeepEqual(got, want) {
		t.Errorf("client.Draws.PropoLatest(%q) \nhave: %#v\nwant: %#v", game, got, want)
	}
}

func TestDrawService_PropoLatest_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/proposat/last.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something broke", 500)
	})

	var game PropoGame = PropoSat
	_, resp, err := client.Draws.PropoLatest(game)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := resp.StatusCode, 500; got != want {
		t.Errorf("resp status code = %d, want %d", got, want)
	}
}

func TestDrawService_PropoByNumber(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/proposat/201751.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"draw":{"drawTime":"23-12-2017T16:00:00","drawNo":201751,"results":["2","2","1","X","X","1","X","2","1","1","1","X","2","2"]}}`)
	})

	var game PropoGame = PropoSat
	var number = 201751
	d, _, err := client.Draws.PropoByNumber(game, number)
	if err != nil {
		t.Fatal("client.Draws.PropoByNumber returned err:", err)
	}
	want := &PropoDraw{DrawTime: "23-12-2017T16:00:00", DrawNo: 201751, Results: []string{"2", "2", "1", "X", "X", "1", "X", "2", "1", "1", "1", "X", "2", "2"}}
	if got := d; !reflect.DeepEqual(got, want) {
		t.Errorf("client.Draws.PropoByNumber(%q, %d) \nhave: %#v\nwant: %#v", game, number, got, want)
	}
}

func TestDrawService_PropoByNumber_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/proposat/201751.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something broke", 500)
	})

	var game PropoGame = PropoSat
	var number = 201751
	_, resp, err := client.Draws.PropoByNumber(game, number)
	if err == nil {
		t.Error("expected error")
	}
	if got, want := resp.StatusCode, 500; got != want {
		t.Errorf("resp status code = %d, want %d", got, want)
	}
}

func TestDrawService_PropoByDate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/proposat/drawDate/23-12-2017.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"draws":{"draw":[{"drawTime":"23-12-2017T16:00:00","drawNo":201751,"results":["2","2","1","X","X","1","X","2","1","1","1","X","2","2"]}]}}`)
	})

	var game PropoGame = PropoSat
	day, month, year := 23, 12, 2017
	d, _, err := client.Draws.PropoByDate(game, day, month, year)
	if err != nil {
		t.Fatal("client.Draws.PropoByDate returned err:", err)
	}
	want := []PropoDraw{{DrawTime: "23-12-2017T16:00:00", DrawNo: 201751, Results: []string{"2", "2", "1", "X", "X", "1", "X", "2", "1", "1", "1", "X", "2", "2"}}}
	if got := d; !reflect.DeepEqual(got, want) {
		t.Errorf("client.Draws.PropoByDate(%q, %d, %d, %d) \nhave: %#v\nwant: %#v", game, day, month, year, got, want)
	}
}

func TestDrawService_PropoByDate_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/"+defaultDrawsEndpoint+"/proposat/drawDate/23-12-2017.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something broke", 500)
	})

	var game PropoGame = PropoSat
	_, resp, err := client.Draws.PropoByDate(game, 23, 12, 2017)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := resp.StatusCode, 500; got != want {
		t.Errorf("resp status code = %d, want %d", got, want)
	}
}
