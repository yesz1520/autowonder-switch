"""Check bilingual documentation, or record one pair after translation review."""

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path
from urllib.parse import unquote, urlsplit

ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"


def pairs():
    sources = {ROOT / "README.md"}
    for path in DOCS.rglob("*"):
        if not path.is_file():
            continue
        for suffix in (".zh.md", ".i18n.yaml", ".md"):
            if path.name.endswith(suffix):
                sources.add(path.with_name(path.name[:-len(suffix)] + ".md"))
                break
    return sources


def links(path, line):
    prose = re.sub(r"(`+).*?\1", "", line)
    for match in re.finditer(r'\]\((?:<([^>]+)>|([^\s)]+))(?:\s+"[^"]*")?\)', prose):
        url = match[1] or match[2]
        parts = urlsplit(url)
        if parts.scheme or parts.netloc:
            yield ("external", url)
            continue
        target = unquote(parts.path)
        if target and not (path.parent / target).exists():
            raise ValueError(f"missing link target {url}")
        if target and (path.parent / target).resolve() == ROOT / "README.en.md":
            target = target[:-len("README.en.md")] + "README.md"
        elif target.endswith(".zh.md"):
            target = target[:-6] + ".md"
        yield ("link", target, parts.query, parts.fragment)


def signature(path, switch):
    lines = path.read_text(encoding="utf-8").splitlines()
    if len(lines) < 3 or not lines[0].startswith("# ") or lines[1:3] != ["", switch]:
        raise ValueError(f"{path.name}: expected language switch below title: {switch}")
    shape = []
    content = []
    fence = ""
    for number, line in enumerate(lines[3:], 4):
        if fence:
            if re.fullmatch(r" {0,3}" + fence + fence[0] + r"*\s*", line):
                body = "\n".join(content)
                if language == "json":
                    try:
                        json.loads(body)
                    except ValueError as err:
                        raise ValueError(f"{path.name}:{number}: invalid JSON: {err}") from err
                shape.append(("example", language, body))
                fence = ""
            else:
                content.append(line)
        elif opening := re.fullmatch(r" {0,3}(`{3,}|~{3,})(.*)", line):
            fence, language = opening[1], opening[2].strip()
            content = []
        else:
            shape.extend(links(path, line))
            if heading := re.match(r"^(#{1,6}) ", line):
                shape.append(("heading", len(heading[1])))
            elif anchor := re.fullmatch(r'<a id="([^"]+)"></a>', line.strip()):
                shape.append(("anchor", anchor[1]))
            elif line.startswith("|"):
                shape.append(("table", len(re.findall(r"(?<!\\)\|", line))))
    if fence:
        raise ValueError(f"{path.name}: unclosed code fence")
    return shape


def check(source, record):
    if source == ROOT / "README.md":
        en, zh = ROOT / "README.en.md", source
    else:
        en, zh = source, source.with_suffix(".zh.md")
    meta = source.with_suffix(".i18n.yaml")
    if signature(en, f"English | [中文]({zh.name})") != signature(zh, f"[English]({en.name}) | 中文"):
        raise ValueError("headings, anchors, tables, links or fenced examples differ")
    rows = []
    for path in (en, zh):
        digest = subprocess.check_output(["git", "hash-object", str(path)], cwd=ROOT, text=True).strip()
        rows.append(f"{path.name}: {digest}\n")
    expected = "".join(rows)
    if record:
        meta.write_text(expected, encoding="utf-8")
    elif meta.read_text(encoding="utf-8") != expected:
        raise ValueError(f"{meta.name}: review both texts before updating the record")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--record", type=Path, metavar="DOCUMENT.md")
    args = parser.parse_args()
    sources = pairs()
    if args.record:
        source = args.record.resolve()
        if source == ROOT / "README.en.md":
            source = ROOT / "README.md"
        if source not in sources:
            parser.error("--record requires the root README or an English document under docs/")
        sources = {source}
    failed = False
    for source in sorted(sources):
        try:
            check(source, bool(args.record))
        except (OSError, ValueError, subprocess.CalledProcessError) as err:
            print(f"{source.relative_to(ROOT)}: {err}", file=sys.stderr)
            failed = True
    if not failed:
        print(f"Checked {len(sources)} documentation pair(s).")
    return int(failed)


if __name__ == "__main__":
    sys.exit(main())
