import $ from 'jquery';
import {alertError, momentIsAfter} from '../utils.js';

export function ajaxLoadSpec(specId, loadBlocks = true, cached = null) {
	let params = {
		specId,
		loadBlocks,
	};
	if (loadBlocks && cached) {
		// Request only to load blocks if updated since latest cache time
		params.cacheTime = momentIsAfter(cached.updated, cached.blocksUpdated)
			? cached.updated : cached.blocksUpdated;
	}
	return $.get('/ajax/spec', params).fail(alertError);
}

export function ajaxCreateSpec(name, desc, isPublic) {
	return $.post('/ajax/spec/create-spec', {
		name, // must be provided
		desc,
		isPublic,
	}).fail(alertError);
}

export function ajaxSaveSpec(specId, name, desc, isPublic) {
	return $.post('/ajax/spec/save-spec', {
		specId,
		name, // must be provided
		desc,
		isPublic,
	}).fail(alertError);
}

export function ajaxDeleteSpec(specId) {
	return $.post('/ajax/spec/delete-spec', {
		specId,
	}).fail(alertError);
}

export function ajaxLoadSubspecs(specId) {
	return $.get('/ajax/spec/subspecs', {
		specId,
	}).fail(alertError);
}

export function ajaxLoadSubspec(specId, subspecId, loadBlocks = true, cached = null) {
	let params = {
		specId,
		subspecId,
		loadBlocks,
	};
	if (loadBlocks && cached) {
		// Request only to load blocks if updated since latest cache time
		params.cacheTime = momentIsAfter(cached.updated, cached.blocksUpdated)
			? cached.updated : cached.blocksUpdated;
	}
	return $.get('/ajax/spec/subspec', params).fail(alertError);
}

export function ajaxCreateSubspec(specId, name, desc) {
	return $.post('/ajax/spec/create-subspec', {
		specId, // required
		name, // required
		desc,
	}).fail(alertError);
}

export function ajaxSaveSubspec(subspecId, name, desc) {
	return $.post('/ajax/spec/save-subspec', {
		subspecId, // required
		name, // required
		desc,
	}).fail(alertError);
}

export function ajaxDeleteSubspec(subspecId) {
	return $.post('/ajax/spec/delete-subspec', {
		subspecId,
	}).fail(alertError);
}

export function ajaxCreateBlock(specId, subspecId, parentId, insertBeforeId,
		styleType, contentType, title, body, refType, refFields) {
	refFields = refFields || {};
	return $.post('/ajax/spec/create-block', {
		specId, // must be provided
		subspecId, // null if spec-level
		parentId, // null if no parent
		insertBeforeId, // null for append
		styleType, // must be provided
		contentType, // may be null
		title, // may be null or empty string
		body, // may be null or empty string
		refType, // may be null
		...refFields,
	}).fail(alertError);
}

export function ajaxSaveBlock(specId, blockId,
		styleType, contentType, title, body, refType, refFields) {
	refFields = refFields || {};
	return $.post('/ajax/spec/save-block', {
		specId, // must be provided
		blockId, // must be provided
		styleType, // must be provided
		contentType, // may be null
		title, // may be null or empty string
		body, // may be null or empty string
		refType, // may be null
		...refFields,
	}).fail(alertError);
}

export function ajaxMoveBlock(blockId, subspecId, parentId, insertBeforeId) {
	return $.post('/ajax/spec/move-block', {
		blockId, // must be provided
		subspecId, // null if spec-level
		parentId, // null if no parent
		insertBeforeId, // null to insert at end
	}).fail(alertError);
}

export function ajaxDeleteBlock(blockId) {
	return $.post('/ajax/spec/delete-block', {
		blockId,
	}).fail(alertError);
}

export function ajaxLoadUrls(specId) {
	return $.get('/ajax/spec/urls', {
		specId,
	}).fail(alertError);
}

export function ajaxCreateUrl(specId, url) {
	return $.post('/ajax/spec/create-url', {
		specId,
		url,
	}).fail(alertError);
}

export function ajaxRefreshUrl(id, url) {
	return $.post('/ajax/spec/refresh-url', {
		id,
		url,
	}).fail(alertError);
}

export function ajaxDeleteUrl(id) {
	return $.post('/ajax/spec/delete-url', {
		id,
	}).fail(alertError);
}

export function fetchUrlPreview(url) {
	return $.get('/ajax/fetch-url', {url}); // No alert on error
}
