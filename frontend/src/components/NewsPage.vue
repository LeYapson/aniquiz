<template>
  <div class="news-page">
    <div class="news-inner">
      <div class="news-page-header">
        <h1>Journal de bord</h1>
        <p>Suivez l'évolution d'AniQuiz — nouvelles fonctionnalités, correctifs et annonces.</p>
      </div>

      <div class="news-list">
        <article v-for="item in news" :key="item.id" class="news-article">
          <div class="article-meta">
            <span class="article-tag" :class="`tag-${tagColor(item.tag)}`">{{ item.tag }}</span>
            <time :datetime="item.datetime" class="article-date">{{ item.date }}</time>
          </div>

          <h2 class="article-title">{{ item.title }}</h2>

          <div class="article-body markdown-body" v-html="item.bodyHtml"></div>
        </article>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useNews } from '../composables/useNews';

const { news } = useNews();

const TAG_COLORS = { Feature: 'green', Annonce: 'blue', Fix: 'orange' };
const tagColor = (tag) => TAG_COLORS[tag] ?? 'orange';
</script>

<style scoped>
.news-page {
  flex: 1;
  background: #0f0f23;
  min-height: calc(100vh - 64px);
  padding: 40px 24px 80px;
}

.news-inner {
  max-width: 720px;
  margin: 0 auto;
}

/* ── En-tête ── */
.news-page-header {
  margin-bottom: 48px;
  padding-bottom: 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
}
.news-page-header h1 {
  font-size: 2rem;
  font-weight: 800;
  color: #f1f5f9;
  margin: 0 0 8px;
}
.news-page-header p {
  color: #64748b;
  font-size: 0.95rem;
  margin: 0;
}

/* ── Liste d'articles ── */
.news-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.news-article {
  padding: 36px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
}
.news-article:last-child {
  border-bottom: none;
}

/* ── Meta ── */
.article-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.article-tag {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 3px 10px;
  border-radius: 20px;
}
.tag-green  { background: rgba(34, 197, 94, 0.15);  color: #4ade80; }
.tag-blue   { background: rgba(59, 130, 246, 0.15);  color: #60a5fa; }
.tag-orange { background: rgba(249, 115, 22, 0.15);  color: #fb923c; }
.article-date {
  font-size: 0.78rem;
  color: #475569;
}

/* ── Titre ── */
.article-title {
  font-size: 1.35rem;
  font-weight: 800;
  color: #f1f5f9;
  margin: 0 0 16px;
  line-height: 1.3;
}

/* ── Corps Markdown ── */
.markdown-body {
  color: #94a3b8;
  font-size: 0.95rem;
  line-height: 1.7;
}
.markdown-body :deep(p)        { margin: 0 0 12px; }
.markdown-body :deep(p:last-child) { margin-bottom: 0; }
.markdown-body :deep(strong)   { color: #e2e8f0; font-weight: 700; }
.markdown-body :deep(em)       { color: #cbd5e1; }
.markdown-body :deep(a)        { color: #f97316; text-decoration: underline; text-underline-offset: 3px; }
.markdown-body :deep(a:hover)  { color: #fb923c; }
.markdown-body :deep(ul),
.markdown-body :deep(ol)       { padding-left: 1.4em; margin: 0 0 12px; }
.markdown-body :deep(li)       { margin-bottom: 4px; }
.markdown-body :deep(code)     { background: rgba(255,255,255,0.07); padding: 2px 6px; border-radius: 4px; font-size: 0.85em; color: #f97316; }
.markdown-body :deep(pre)      { background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.08); border-radius: 8px; padding: 16px; overflow-x: auto; margin: 0 0 12px; }
.markdown-body :deep(pre code) { background: none; padding: 0; color: #e2e8f0; font-size: 0.88em; }
.markdown-body :deep(hr)       { border: none; border-top: 1px solid rgba(255,255,255,0.07); margin: 16px 0; }
.markdown-body :deep(h3)       { color: #e2e8f0; font-size: 1rem; font-weight: 700; margin: 16px 0 8px; }

/* ── Responsive ── */
@media (max-width: 600px) {
  .news-page { padding: 24px 16px 60px; }
  .news-page-header h1 { font-size: 1.5rem; }
  .article-title { font-size: 1.15rem; }
}
</style>
