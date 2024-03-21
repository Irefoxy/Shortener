package short_url_generator

import "testing"

func TestParser(t *testing.T) {
	tests := []struct {
		name          string
		original      string
		expectedShort string
	}{
		{"yandex", "yandex.com", "YtyeWuJ_tKg="},
		{"short", "a", "MwKEdy5lKwU="},
		{"long", "https://market.yandex.ru/product--iphone-15-pro-max/1912857921?sku=102254215793&do-waremd5=v3h8jWdUJhCRBbeGKBExvA&uniqueId=77942856&resale_goods=resale_new", "ofcvP9u98Fo="},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := New()
			short, err := parser.Generate(test.original)
			if err != nil {
				t.Fatal(err)
			}
			if short != test.expectedShort {
				t.Errorf("Expected %s: %s, but got %s", test.original, test.expectedShort, short)
			}
		})
	}
}
