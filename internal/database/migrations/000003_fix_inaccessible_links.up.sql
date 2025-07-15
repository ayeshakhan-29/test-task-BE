-- Update existing records where inaccessible_links is a number to be an empty array
UPDATE crawl_results 
SET inaccessible_links = '[]' 
WHERE JSON_TYPE(inaccessible_links) = 'INTEGER';
