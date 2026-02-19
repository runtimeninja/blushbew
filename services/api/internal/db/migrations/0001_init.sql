-- Blog posts
create table if not exists blog_posts (
  id bigserial primary key,
  slug text not null unique,
  title text not null,
  excerpt text not null default '',
  content_md text not null,
  category text not null default 'general',
  status text not null default 'draft', -- draft|published
  published_at timestamptz null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_blog_posts_status_published_at
on blog_posts(status, published_at desc);

-- Diagnosis requests (AI tool usage)
create table if not exists diagnosis_requests (
  id bigserial primary key,
  problem_text text not null,
  skin_type text not null default '',
  climate text not null default '',
  event_type text not null default '',
  result_json jsonb not null,
  created_at timestamptz not null default now()
);

create index if not exists idx_diagnosis_requests_created_at
on diagnosis_requests(created_at desc);

-- Analytics events (simple event log)
create table if not exists analytics_events (
  id bigserial primary key,
  event_name text not null,
  payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_analytics_events_created_at
on analytics_events(created_at desc);

create index if not exists idx_analytics_events_event_name
on analytics_events(event_name);

-- Admin users (simple)
create table if not exists admin_users (
  id bigserial primary key,
  email text not null unique,
  password_hash text not null,
  created_at timestamptz not null default now()
);
