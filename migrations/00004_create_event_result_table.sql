-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS event_result (
    id INTEGER PRIMARY KEY AUTOINCREMENT, 
    event_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    best_lap_time INTEGER NOT NULL,
    average_lap_time INTEGER NOT NULL,
    position INTEGER NOT NULL,
    number_of_laps INTEGER NOT NULL,
    FOREIGN KEY (event_id) REFERENCES event (id),
    FOREIGN KEY (user_id) REFERENCES user (id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE event_result;
-- +goose StatementEnd
