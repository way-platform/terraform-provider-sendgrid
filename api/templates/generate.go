package templates

//go:generate echo "--- overlay ---"
//go:generate sh -c "openapi-overlay apply overlay.yaml 01-original.json > 02-overlayed.json"
//go:generate echo "--- downconvert (3.1 → 3.0) ---"
//go:generate npx @apiture/openapi-down-convert@0.14.1 --input 02-overlayed.json --output 03-downconverted.json
//go:generate echo "--- oapi-codegen ---"
//go:generate oapi-codegen -config config.yaml 03-downconverted.json
