ALTER TABLE course_reviews
    ADD CONSTRAINT course_reviews_comment_length_check
    CHECK (comment IS NULL OR char_length(comment) <= 2000);

ALTER TABLE content_reviews
    ADD CONSTRAINT content_reviews_comment_length_check
    CHECK (comment IS NULL OR char_length(comment) <= 2000);
