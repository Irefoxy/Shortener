package converters

import (
	m "Yandex/internal/api/gin_api/models"
	"Yandex/internal/models"
)

func ApiUrlToEntry(url, uuid string) []models.Entry {
	return []models.Entry{{
		Id:          uuid,
		OriginalUrl: url,
	}}
}

func ApiShortUrlsToEntry(uuid string, urls ...string) (result []models.Entry) {
	for _, url := range urls {
		result = append(result, models.Entry{
			Id:       uuid,
			ShortUrl: url,
		})
	}
	return result
}

func EntryToApiUrl(entry models.Entry, targetAddress string) string {
	return targetAddress + "/" + entry.ShortUrl
}

func ApiJSONUrlToEntry(url m.URL, uuid string) []models.Entry {
	return []models.Entry{{
		Id:          uuid,
		OriginalUrl: url.Url,
	}}
}

func EntryToApiJSONUrl(entry models.Entry, targetAddress string) m.ShortURL {
	return m.ShortURL{Result: targetAddress + "/" + entry.ShortUrl}
}

func ApiJSONUrlBatchToEntry(urls []m.BatchURL, uuid string) (entries []models.Entry) {
	for _, url := range urls {
		entries = append(entries, models.Entry{
			Id:          uuid,
			OriginalUrl: url.Original,
		})
	}
	return
}

func EntryToApiJSONUrlBatch(entries []models.Entry, targetAddress string, requests []m.BatchURL) (result []m.BatchShortURL) {
	for i, entry := range entries {
		result = append(result, m.BatchShortURL{
			Id:    requests[i].Id,
			Short: targetAddress + "/" + entry.ShortUrl,
		})
	}
	return
}
