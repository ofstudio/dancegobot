/*
AS IS: events.data = '{
  "id": "123abc",
  "text": "Event announcement",
  "message_id": "ABC123"
  "singles": [
    {
      "single_signup": true,
      ...
    }
  ]
  ...
}'

TO BE: events.data = '{
  "id": "123abc",
  "caption": "Event announcement",
  post: {
    "inline_message_id": "ABC123"
  }
  "singles": [
    {
      "as_single": true,
      ...
    }
  ]
  ...
}'

*/


-- Step 1: Add the post object field
UPDATE events
SET data = json_remove(
        json_set(data, '$.post', json_object('inline_message_id', json_extract(data, '$.message_id'))),
        '$.message_id'
           )
WHERE json_extract(data, '$.message_id') IS NOT NULL;


-- Step 2: Rename the text field to caption
UPDATE events
SET data = json_remove(
        json_set(data, '$.caption', json_extract(data, '$.text')),
        '$.text')
WHERE json_extract(data, '$.text') IS NOT NULL;


-- Step 3: Rename the single_signup field to as_single
UPDATE events
SET data = REPLACE (data, '"single_signup"', '"as_single"')
WHERE data IS NOT NULL;
