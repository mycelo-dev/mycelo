# ─────────────────────────────────────────────
# LOAD MODELS  (done once at startup)
# ─────────────────────────────────────────────

print("Loading models — this takes ~30 seconds on first run...")

nlp = spacy.load("en_core_web_sm")

ner_model = GLiNER.from_pretrained("urchade/gliner_large-v2.1")

_re_tokenizer = AutoTokenizer.from_pretrained("Babelscape/rebel-large")
_re_model_raw = AutoModelForSeq2SeqLM.from_pretrained("Babelscape/rebel-large")

def re_model(text: str) -> list[dict]:
    inputs  = _re_tokenizer(text, return_tensors="pt", truncation=True, max_length=512)
    outputs = _re_model_raw.generate(**inputs, max_length=512)
    decoded = _re_tokenizer.decode(outputs[0], skip_special_tokens=False)
    return [{"generated_text": decoded}]

embedder = SentenceTransformer("all-MiniLM-L6-v2")

print("Models loaded.\n")