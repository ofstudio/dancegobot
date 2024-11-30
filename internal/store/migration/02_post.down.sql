UPDATE events
SET data = json_set(data, '$.message_id', json_extract(data, '$.post.inline_message_id'))
WHERE json_extract(data, '$.post.inline_message_id') IS NOT NULL;

UPDATE events
SET data = json_remove(data, '$.post')
WHERE json_extract(data, '$.post') IS NOT NULL;
