package main

import (
	"strings"
	"testing"
)

func TestExtractDesktopWidgetDocumentMissing(t *testing.T) {
	doc, ok, err := extractDesktopWidgetDocument([]byte(`{"apps":[]}`))
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if ok {
		t.Fatal("want ok=false for missing desktopWidgets")
	}
	if doc.Version != 0 || doc.Widgets != nil {
		t.Errorf("want zero doc, got %+v", doc)
	}
}

func TestExtractDesktopWidgetDocumentNullAndEmpty(t *testing.T) {
	for _, data := range []string{
		`{"desktopWidgets":null}`,
		`{"desktopWidgets": null }`,
	} {
		doc, ok, err := extractDesktopWidgetDocument([]byte(data))
		if err != nil {
			t.Fatalf("input %s: want nil err, got %v", data, err)
		}
		if ok {
			t.Fatalf("input %s: want ok=false", data)
		}
		if doc.Version != 0 {
			t.Errorf("input %s: want zero doc, got %+v", data, doc)
		}
	}
}

func TestExtractDesktopWidgetDocumentValid(t *testing.T) {
	data := []byte(`{"desktopWidgets":{"version":1,"widgets":{},"notes":{},"reminders":{},"timers":{},"stopwatches":{},"stopwatchLaps":{},"weatherCache":{},"notificationDeliveries":{},"protectedSecrets":{}}}`)
	doc, ok, err := extractDesktopWidgetDocument(data)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if !ok {
		t.Fatal("want ok=true")
	}
	if doc.Version != 1 {
		t.Errorf("doc.Version = %d, want 1", doc.Version)
	}
}

func TestExtractDesktopWidgetDocumentDecodeError(t *testing.T) {
	doc, ok, err := extractDesktopWidgetDocument([]byte(`{"desktopWidgets":"not-json"}`))
	if err == nil {
		t.Fatal("want err for invalid inner JSON")
	}
	if ok {
		t.Fatal("want ok=false on decode error")
	}
	if doc.Version != 0 {
		t.Errorf("want zero doc, got %+v", doc)
	}
	if !strings.Contains(err.Error(), "DESKTOP_WIDGET_STORE_INVALID") {
		t.Errorf("err = %q, want DESKTOP_WIDGET_STORE_INVALID prefix", err.Error())
	}
}

func TestLauncherConfigDocumentIsWidgetLibrary(t *testing.T) {
	cases := []struct {
		data string
		want bool
	}{
		{`{"widgets":{}}`, true},
		{`{"widgets":{},"initialized":true}`, false},
		{`{"widgets":{},"desktopWidgets":{}}`, false},
		{`{"widgets":{},"twoFactor":{}}`, false},
		{`{"apps":[]}`, false},
		{`not-json`, false},
		{`{"widgets":{},"unknownKey":1}`, true},
	}
	for _, c := range cases {
		got := launcherConfigDocumentIsWidgetLibrary([]byte(c.data))
		if got != c.want {
			t.Errorf("input %s: got %v, want %v", c.data, got, c.want)
		}
	}
}

func TestPortableDesktopWidgetDocumentClearsDeviceFields(t *testing.T) {
	doc := DesktopWidgetDocument{
		Version:   1,
		Revision:  7,
		UpdatedAt: "2026-09-30T00:00:00Z",
		Widgets: map[string]DesktopWidget{
			"w1": {ID: "w1", Type: "clock", Title: "Clock", Revision: 1},
		},
		WeatherCache: map[string]DesktopWeatherSnapshot{
			"w1": {WidgetID: "w1"},
		},
		NotificationDeliveries: map[string]DesktopNotificationRecord{
			"k1": {DeliveryKey: "k1"},
		},
		ProtectedSecrets: map[string]DesktopProtectedSecret{
			"k2": {Provider: "p"},
		},
	}
	out := portableDesktopWidgetDocument(doc)
	if len(out.WeatherCache) != 0 {
		t.Errorf("WeatherCache len = %d, want 0", len(out.WeatherCache))
	}
	if len(out.NotificationDeliveries) != 0 {
		t.Errorf("NotificationDeliveries len = %d, want 0", len(out.NotificationDeliveries))
	}
	if len(out.ProtectedSecrets) != 0 {
		t.Errorf("ProtectedSecrets len = %d, want 0", len(out.ProtectedSecrets))
	}
	if out.Version != 1 || out.Revision != 7 {
		t.Errorf("scalar fields not preserved: %+v", out)
	}
	if _, ok := out.Widgets["w1"]; !ok {
		t.Errorf("Widgets not preserved: %+v", out.Widgets)
	}
}

func TestDesktopWidgetDocumentForImportKeepsSecrets(t *testing.T) {
	doc := DesktopWidgetDocument{
		Version: 1,
		WeatherCache: map[string]DesktopWeatherSnapshot{
			"w1": {WidgetID: "w1"},
		},
		ProtectedSecrets: map[string]DesktopProtectedSecret{
			"k1": {Provider: "p", ProtectedBlob: "blob"},
		},
	}
	out := desktopWidgetDocumentForImport(doc)
	if len(out.WeatherCache) != 0 {
		t.Errorf("WeatherCache len = %d, want 0", len(out.WeatherCache))
	}
	if _, ok := out.ProtectedSecrets["k1"]; !ok {
		t.Errorf("ProtectedSecrets not preserved: %+v", out.ProtectedSecrets)
	}
}
