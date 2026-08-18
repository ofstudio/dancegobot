CREATE TABLE "subscriptions"
(
    "subscriber_id" INTEGER   NOT NULL,
    "chat_id"       INTEGER   NOT NULL,
    "data"          JSON      NOT NULL,
    "created_at"    TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    PRIMARY KEY (subscriber_id, chat_id)
);

CREATE INDEX "subscriptions_chat_id_created_at"
    ON "subscriptions" ("chat_id", "created_at", "subscriber_id");

-- Existing posts must not produce notifications after the feature is deployed.
UPDATE events
SET data = json_set(data, '$.subscribers_notified', json('true'))
WHERE json_extract(data, '$.post.chat_message_id') IS NOT NULL;
