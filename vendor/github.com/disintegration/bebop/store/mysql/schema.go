package mysql

var migrate = []string{
	`
		create table if not exists users (
			id            bigint            not null auto_increment,
			name          varchar(50)       default null,
			created_at    datetime(6)       not null,
			auth_service  varchar(50)       not null,
			auth_id       varchar(50)       not null,
			blocked       boolean           not null default false,
			admin         boolean           not null default false,
			avatar        varchar(50)       not null default '',

			primary key (id),
			unique index (name),
			unique index (auth_service, auth_id)
		) default charset = utf8mb4;
	`,
	`
		create table if not exists topics (
			id               bigint        not null auto_increment,
			author_id        bigint        not null references users(id),
			title            varchar(200)  not null,
			created_at       datetime(6)   not null,
			last_comment_at  datetime(6)   not null,
			deleted          boolean       not null default false,
			comment_count    int           not null default 0,

			primary key (id),
			index (last_comment_at)
		) default charset = utf8mb4;
	`,
	`
		create table if not exists comments (
			id          bigint       not null auto_increment,
			topic_id    bigint       not null references topics(id),
			author_id   bigint       not null references users(id),
			content     text         not null,
			created_at  datetime(6)  not null,
			deleted     boolean      not null default false,

			primary key (id),
			index (topic_id),
			index (created_at)
		) default charset = utf8mb4;
	`,
}

var drop = []string{
	`drop table if exists users cascade`,
	`drop table if exists topics cascade`,
	`drop table if exists comments cascade`,
}
