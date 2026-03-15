from transformers import AutoTokenizer, AutoModelForSeq2SeqLM

_re_tokenizer = AutoTokenizer.from_pretrained("Babelscape/rebel-large")
_re_model_raw = AutoModelForSeq2SeqLM.from_pretrained("Babelscape/rebel-large")


def re_model(text: str) -> list[dict]:
    inputs  = _re_tokenizer(text, return_tensors="pt", truncation=True, max_length=512)
    outputs = _re_model_raw.generate(**inputs, max_length=512)
    decoded = _re_tokenizer.decode(outputs[0], skip_special_tokens=False)
    return [{"generated_text": decoded}]


def parse_rebel_output(text: str) -> list[dict]:
    """
    REBEL format: <triplet> HEAD <subj> TAIL <obj> RELATION
    """
    triplets = []
    current  = {}
    mode     = None

    for token in text.replace("<s>", "").replace("<pad>", "").split():
        if token == "<triplet>":
            if current and "_buf" in current and mode:
                current[mode] = current["_buf"].strip()
            if current:
                triplets.append(current)
            current = {"_buf": ""}
            mode = "head"
        elif token == "<subj>":
            current["head"] = current.get("_buf", "").strip()
            current["_buf"] = ""
            mode = "tail"
        elif token == "<obj>":
            current["tail"] = current.get("_buf", "").strip()
            current["_buf"] = ""
            mode = "relation"
        else:
            current["_buf"] = current.get("_buf", "") + " " + token

    if current and "_buf" in current and mode:
        current[mode] = current["_buf"].strip()
    if current:
        triplets.append(current)

    return [
        {
            "head":     t["head"],
            "relation": t["relation"].replace("</s>", "").strip(),
            "tail":     t["tail"],
        }
        for t in triplets
        if t.get("head") and t.get("relation") and t.get("tail")
    ]
