package postgresql

var migrate = []string{
	`
		create table if not exists users (
			id            bigserial    not null primary key,
			name          text         default null,
			created_at    timestamptz  not null,
			auth_service  text         not null,
			auth_id       text         not null,
			blocked       boolean      not null default false,
			admin         boolean      not null default false,
			avatar        text         not null default ''
		);
		create unique index on users(lower(name));
		create unique index on users(auth_service, auth_id);
	`,
	`
		create table if not exists topics (
			id               bigserial    not null primary key,
			author_id        bigint       not null references users(id),
			title            text         not null,
			created_at       timestamptz  not null,
			last_comment_at  timestamptz  not null,
			deleted          boolean      not null default false,
			comment_count    int          not null default 0
		);
		create index on topics(last_comment_at);
	`,
	`
		create table if not exists comments (
			id          bigserial     not null primary key,
			topic_id    bigint        not null references topics(id),
			author_id   bigint        not null references users(id),
			content     text          not null,
			created_at  timestamptz   not null,
			deleted     boolean       not null default false
		);
		create index on comments(topic_id);
		create index on comments(created_at);
	`,
}

var drop = []string{
	`drop table if exists users cascade`,
	`drop table if exists topics cascade`,
	`drop table if exists comments cascade`,
}
