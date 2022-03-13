## Gemserver
Gemserver is a Gemini search engine.

### Generating self-signed cert
```bash
$ openssl req -new -subj "/C=IN/ST=Delhi/CN=localhost" -newkey rsa:4096 -nodes -keyout gemini.key -out gemini.csr

$ openssl x509 -req -days 365 -in gemini.csr -signkey gemini.key -out gemini.crt
```
You can change the `CN=localhost` in `-subj` to point to a different domain.