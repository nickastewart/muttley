-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS event (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  location_id INTEGER NOT NULL,
  type VARCHAR(100) NOT NULL,
  date TEXT NOT NULL,
  total_drivers INTEGER NOT NULL,
  FOREIGN KEY (location_id) REFERENCES location (id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE event;
-- +goose StatementEnd
