<section class="hero">
  <div class="hero-text">
    <p class="eyebrow">Penguin Console</p>
    <h1>会社と工事を、すばやく見渡す。</h1>
    <p class="lead">
      会社一覧と工事一覧を中心に、日常の確認と更新を最短で行うための
      シンプルな管理ビューです。
    </p>
    <div class="actions">
      <a class="primary" href="/companies">会社一覧</a>
      <a class="ghost" href="/kojies">工事一覧</a>
    </div>
  </div>
  <div class="hero-panel">
    <div class="panel-row">
      <span>現在の業務</span>
      <strong>会社・工事の把握</strong>
    </div>
    <div class="panel-row">
      <span>更新フロー</span>
      <strong>一覧 → 編集 → 保存</strong>
    </div>
  </div>
</section>

<section class="quick-links">
  <a href="/companies">
    <span>Company</span>
    <strong>会社一覧を開く</strong>
  </a>
  <a href="/kojies">
    <span>Koji</span>
    <strong>工事一覧を開く</strong>
  </a>
</section>

<style>
  :global(body) {
    background: radial-gradient(circle at top, #f1f5f9, #e2e8f0);
    color: #0f172a;
  }

  .hero {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 24px;
    padding: 48px 20px 24px;
    max-width: 1000px;
    margin: 0 auto;
  }

  .hero-text {
    display: grid;
    gap: 12px;
  }

  .eyebrow {
    text-transform: uppercase;
    letter-spacing: 0.2em;
    font-size: 0.7rem;
    color: #64748b;
    margin: 0;
  }

  h1 {
    font-size: clamp(2rem, 3vw, 2.6rem);
    margin: 0;
  }

  .lead {
    margin: 0;
    color: #475569;
    line-height: 1.6;
  }

  .actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: 8px;
  }

  .primary,
  .ghost {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 10px 18px;
    border-radius: 999px;
    text-decoration: none;
    font-weight: 600;
  }

  .primary {
    background: #0f172a;
    color: #ffffff;
  }

  .ghost {
    background: #e2e8f0;
    color: #0f172a;
  }

  .hero-panel {
    background: #0f172a;
    color: #f8fafc;
    border-radius: 18px;
    padding: 18px 20px;
    display: grid;
    gap: 16px;
    box-shadow: 0 18px 36px rgba(15, 23, 42, 0.2);
  }

  .panel-row span {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.16em;
    color: rgba(248, 250, 252, 0.7);
  }

  .panel-row strong {
    display: block;
    font-size: 1.1rem;
    margin-top: 6px;
  }

  .quick-links {
    max-width: 1000px;
    margin: 0 auto 48px;
    padding: 0 20px 32px;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
  }

  .quick-links a {
    display: grid;
    gap: 8px;
    padding: 16px 18px;
    border-radius: 16px;
    background: #ffffff;
    text-decoration: none;
    color: #0f172a;
    box-shadow: 0 12px 24px rgba(15, 23, 42, 0.08);
  }

  .quick-links span {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: #94a3b8;
  }
</style>
