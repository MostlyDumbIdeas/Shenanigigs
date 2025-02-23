package migrations

import "shenanigigs/common/database/schema"

var CreateJobsTable = schema.Migration{
	Version:     1,
	Description: "Create jobs table",
	Up: `
		CREATE TABLE IF NOT EXISTS jobs (
			id UUID,
			title String,
			company String,
			location String,
			description String,
			technologies Array(String),
			experience_level String,
			compensation_min Nullable(Float64),
			compensation_max Nullable(Float64),
			compensation_currency String,
			compensation_period String,
			remote_policy String,
			source String,
			source_url String,
			created_at DateTime,
			updated_at DateTime,
			raw_data String,
			PRIMARY KEY (id)
		) ENGINE = ReplacingMergeTree(updated_at)
		PARTITION BY toYYYYMM(created_at)
		ORDER BY (id, created_at)
		SETTINGS index_granularity = 8192
	`,
	Down: `DROP TABLE IF EXISTS jobs`,
}
