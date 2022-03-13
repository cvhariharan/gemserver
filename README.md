## Gemserver
Gemserver is a [geminispace](https://en.wikipedia.org/wiki/Gemini_(protocol)) search engine. Built using [gemini-server](https://github.com/cvhariharan/gemini-server) and uses [gemini-crawler](https://github.com/cvhariharan/gemini-crawler) to index the geminispace.

### Generating self-signed cert
```bash
$ openssl req -new -subj "/C=IN/ST=Delhi/CN=localhost" -newkey rsa:4096 -nodes -keyout gemini.key -out gemini.csr

$ openssl x509 -req -days 365 -in gemini.csr -signkey gemini.key -out gemini.crt
```
You can change the `CN=localhost` in `-subj` to point to a different domain.