/*
AS IS: events.data = '{
  "id": "123abc",
  "message_id": "ABC123"
  ...
}'

TO BE: events.data = '{
  "id": "123abc",
  post: {
    "inline_message_id": "ABC123"
  }
  ...
}'

*/


-- Step 1: Add the post object field
UPDATE events
SET data = json_set(data, '$.post', json_object('inline_message_id', json_extract(data, '$.message_id')))
WHERE json_extract(data, '$.message_id') IS NOT NULL;

-- Step 2: Remove the message_id field
UPDATE events
SET data = json_remove(data, '$.message_id')
WHERE json_extract(data, '$.message_id') IS NOT NULL;

