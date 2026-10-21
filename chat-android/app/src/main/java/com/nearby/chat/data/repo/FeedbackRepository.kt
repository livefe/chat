package com.nearby.chat.data.repo

import com.nearby.chat.core.model.FeedbackItem
import com.nearby.chat.core.network.ensureSuccess
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.remote.api.FeedbackApi
import com.nearby.chat.data.remote.dto.SubmitFeedbackRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class FeedbackRepository @Inject constructor(
    private val feedbackApi: FeedbackApi,
) {
    suspend fun submit(
        type: Int,
        content: String,
        contact: String?,
        images: List<Long>?,
    ) {
        feedbackApi.submit(
            SubmitFeedbackRequest(type = type, content = content, contact = contact, images = images),
        ).ensureSuccess()
    }

    suspend fun myList(page: Int, pageSize: Int = 10): Pair<Int, List<FeedbackItem>> {
        val data = feedbackApi.myList(page, pageSize).unwrap()
        return data.total to data.list.map { it.toDomain() }
    }

    suspend fun myDetail(id: Long): FeedbackItem = feedbackApi.myDetail(id).unwrap().toDomain()
}
