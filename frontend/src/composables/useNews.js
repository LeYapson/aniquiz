import { marked } from 'marked';

marked.setOptions({ breaks: true });

function parseFrontmatter(raw) {
  const match = raw.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n([\s\S]*)$/);
  if (!match) return { data: {}, content: raw };

  const data = Object.fromEntries(
    match[1]
      .split('\n')
      .filter((line) => line.includes(':'))
      .map((line) => {
        const idx = line.indexOf(':');
        return [
          line.slice(0, idx).trim(),
          line.slice(idx + 1).trim().replace(/^"(.*)"$/, '$1'),
        ];
      }),
  );

  return { data, content: match[2].trim() };
}

const rawFiles = import.meta.glob('../content/news/*.md', {
  query: '?raw',
  import: 'default',
  eager: true,
});

export function useNews() {
  const news = Object.entries(rawFiles)
    .map(([path, raw]) => {
      const { data, content } = parseFrontmatter(raw);
      return {
        id: Number(data.id),
        date: data.date ?? '',
        datetime: data.datetime ?? '',
        tag: data.tag ?? 'Annonce',
        title: data.title ?? '',
        bodyHtml: marked.parse(content),
        slug: path.replace(/.*\//, '').replace(/\.md$/, ''),
      };
    })
    .sort((a, b) => b.datetime.localeCompare(a.datetime));

  return { news };
}
