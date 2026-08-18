DROP TABLE "subscriptions";

UPDATE events
SET data = json_remove(data, '$.subscribers_notified')
WHERE json_extract(data, '$.subscribers_notified') IS NOT NULL;
