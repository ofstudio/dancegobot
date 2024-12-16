UPDATE events
SET data =json_remove(
        json_set(data, '$.message_id', json_extract(data, '$.post.inline_message_id')),
        '$.post')
WHERE json_extract(data, '$.post.inline_message_id') IS NOT NULL;

UPDATE events
SET data = json_remove(
        json_set(data, '$.text', json_extract(data, '$.caption')),
        '$.caption')
WHERE json_extract(data, '$.caption') IS NOT NULL;

UPDATE events
SET data = REPLACE (data, '"as_single"', '"single_signup"')
WHERE data IS NOT NULL;
