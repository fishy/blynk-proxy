GO=go
GCLOUD=gcloud
p=blynk-proxy

.PHONY: deploy tail

deploy:
	$(GO) mod tidy
	$(GCLOUD) app deploy --project=$(p)

tail:
	$(GCLOUD) app logs tail -s default --project=$(p)
