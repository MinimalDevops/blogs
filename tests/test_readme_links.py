import re
from pathlib import Path


def slugify(text: str) -> str:
    """Simplified slug function similar to GitHub's"""
    value = text.lower()
    value = re.sub(r'[^\w\s-]', '', value)
    value = value.replace(' ', '-')
    return value


def extract_anchor_links(text: str):
    pattern = re.compile(r'\[[^\]]+\]\(#([^\)]+)\)')
    return pattern.findall(text)


def extract_heading_slugs(text: str):
    slugs = []
    for line in text.splitlines():
        m = re.match(r'^#+\s+(.*)', line)
        if m:
            slugs.append(slugify(m.group(1)))
    return slugs


def test_markdown_anchor_links_have_headings():
    readme = Path('README.md').read_text(encoding='utf-8')
    anchors = [slugify(a) for a in extract_anchor_links(readme)]
    heading_slugs = set(extract_heading_slugs(readme))
    missing = [a for a in anchors if a not in heading_slugs]
    assert not missing, f"Missing headings for anchors: {missing}"
