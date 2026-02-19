package blog

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Excerpt     string     `json:"excerpt"`
	ContentMD   string     `json:"content_md"`
	Category    string     `json:"category"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) ListPublished(ctx context.Context, limit int) ([]Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		select id, slug, title, excerpt, content_md, category, status, published_at, created_at, updated_at
		from blog_posts
		where status='published'
		order by published_at desc nulls last, created_at desc
		limit $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list published: %w", err)
	}
	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ContentMD, &p.Category, &p.Status, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (r *Repo) GetPublishedBySlug(ctx context.Context, slug string) (*Post, error) {
	var p Post
	err := r.pool.QueryRow(ctx, `
		select id, slug, title, excerpt, content_md, category, status, published_at, created_at, updated_at
		from blog_posts
		where status='published' and slug=$1
		limit 1
	`, slug).Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ContentMD, &p.Category, &p.Status, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) AdminCreate(ctx context.Context, p Post) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		insert into blog_posts(slug, title, excerpt, content_md, category, status)
		values($1,$2,$3,$4,$5,$6)
		returning id
	`, p.Slug, p.Title, p.Excerpt, p.ContentMD, p.Category, p.Status).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create: %w", err)
	}
	return id, nil
}

func (r *Repo) AdminPublish(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `
		update blog_posts
		set status='published', published_at=now(), updated_at=now()
		where id=$1
	`, id)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}

func (r *Repo) AdminList(ctx context.Context, limit int) ([]Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		select id, slug, title, excerpt, content_md, category, status, published_at, created_at, updated_at
		from blog_posts
		order by created_at desc
		limit $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("admin list: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.ContentMD, &p.Category, &p.Status, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}
