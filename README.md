# wmf

This module provides a unified client for working with Wikimedia Foundation APIs. It wraps the MediaWiki Actions API, REST endpoints, Lift Wing inference services, Commons file access, and Wikibase entity lookups behind a single, consistent interface.

Pagination and continuation are handled automatically.

Most functions also support passing additional request options, enabling customization of query parameters or headers. This can be used to adjust limits, namespaces, or other API-specific settings on a per-request basis.

A central client manages configuration such as base URLs, user agent, retries and backoff behavior. The Wikimedia site matrix is loaded and cached to resolve database names to the correct project URLs. Built-in retry logic helps with rate limits and transient upstream errors, and concurrent helpers are available for fetching multiple HTML pages or revisions efficiently.
