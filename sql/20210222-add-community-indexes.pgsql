CREATE INDEX comment_by_spec ON spec_community_comment (spec_id); -- used in community review
CREATE INDEX comment_updated_by_user ON spec_community_comment (user_id, updated_at); -- used on home page and in community review
