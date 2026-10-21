package com.nearby.chat.core.paging
class OffsetPageTracker(val pageSize: Int) {
    var page: Int = 1
        private set

    fun reset() {
        page = 1
    }

    data class Result<T>(
        val list: List<T>,
        val hasMore: Boolean,
    )

    fun <T> apply(
        existing: List<T>,
        total: Int,
        items: List<T>,
        replace: Boolean = page == 1,
        countedSize: (List<T>) -> Int = { it.size },
    ): Result<T> {
        val merged = if (replace) items else existing + items
        val hasMore = countedSize(merged) < total && countedSize(items) >= pageSize
        if (hasMore) page += 1
        return Result(merged, hasMore)
    }
}
