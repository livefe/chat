package com.nearby.chat.data.repo

import com.nearby.chat.core.model.FeedItem
import com.nearby.chat.core.model.PostCheckInCode
import com.nearby.chat.core.model.PostCheckInResult
import com.nearby.chat.core.model.PostDraft
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.model.PostSignupPage
import com.nearby.chat.core.network.ensureSuccess
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.mapper.toDto
import com.nearby.chat.data.remote.api.PostApi
import com.nearby.chat.data.remote.dto.CreatePostRequest
import com.nearby.chat.data.remote.dto.PostCheckInRequest
import com.nearby.chat.data.remote.dto.PostSignupRequest
import com.nearby.chat.data.remote.dto.UpdateSignupRemarkRequest
import com.nearby.chat.data.remote.dto.WithdrawPostRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PostRepository @Inject constructor(
    private val postApi: PostApi,
) {
    suspend fun list(
        page: Int,
        pageSize: Int = 20,
        latitude: Double? = null,
        longitude: Double? = null,
        roomId: Long? = null,
        category: String? = null,
    ): Pair<Int, List<FeedItem>> {
        val data = postApi.list(page, pageSize, latitude, longitude, roomId, category).unwrap()
        return data.total to data.list.mapNotNull { it.toDomain() }
    }

    suspend fun myList(
        page: Int,
        pageSize: Int = 30,
        latitude: Double? = null,
        longitude: Double? = null,
    ): Pair<Int, List<PostItem>> {
        val data = postApi.myList(page, pageSize, latitude, longitude).unwrap()
        return data.total to data.list.map { it.toDomain() }
    }

    suspend fun myJoinedList(
        page: Int,
        pageSize: Int = 30,
        latitude: Double? = null,
        longitude: Double? = null,
    ): Pair<Int, List<PostItem>> {
        val data = postApi.myJoinedList(page, pageSize, latitude, longitude).unwrap()
        return data.total to data.list.map { it.toDomain() }
    }

    suspend fun userList(
        userId: Long,
        page: Int,
        pageSize: Int = 20,
        latitude: Double? = null,
        longitude: Double? = null,
    ): Pair<Int, List<PostItem>> {
        val data = postApi.userList(userId, page, pageSize, latitude, longitude).unwrap()
        return data.total to data.list.map { it.toDomain() }
    }

    suspend fun detail(id: Long): PostItem = postApi.detail(id).unwrap().toDomain()

    suspend fun create(draft: PostDraft): Long = postApi.create(draft.toRequest()).unwrap().id

    suspend fun update(draft: PostDraft) {
        require(draft.id > 0)
        postApi.update(draft.toRequest()).ensureSuccess()
    }

    suspend fun signup(postId: Long): Int = postApi.signup(PostSignupRequest(postId)).unwrap().status

    suspend fun signupCancel(postId: Long) {
        postApi.signupCancel(PostSignupRequest(postId)).ensureSuccess()
    }

    suspend fun signupCheckInCode(postId: Long): PostCheckInCode = postApi.signupCheckInCode(postId).unwrap().toDomain()

    suspend fun signupCheckIn(
        payload: String = "",
        postId: Long = 0,
        userId: Long = 0,
    ): PostCheckInResult = postApi.signupCheckIn(
        PostCheckInRequest(
            payload = payload,
            postId = postId,
            userId = userId,
        ),
    ).unwrap().toDomain()

    suspend fun signupList(
        postId: Long,
        page: Int,
        pageSize: Int = 30,
        status: Int? = null,
        checkIn: Int? = null,
    ): PostSignupPage = postApi.signupList(postId, page, pageSize, status, checkIn).unwrap().toDomain()

    suspend fun updateSignupRemark(postId: Long, userId: Long, remark: String) {
        postApi.updateSignupRemark(
            UpdateSignupRemarkRequest(postId = postId, userId = userId, remark = remark),
        ).ensureSuccess()
    }

    suspend fun withdraw(id: Long) {
        postApi.withdraw(WithdrawPostRequest(id)).ensureSuccess()
    }

    suspend fun deleteMine(id: Long) {
        postApi.deleteMine(WithdrawPostRequest(id)).ensureSuccess()
    }

    suspend fun removeJoined(id: Long) {
        postApi.removeJoined(WithdrawPostRequest(id)).ensureSuccess()
    }

    private fun PostDraft.toRequest(): CreatePostRequest = CreatePostRequest(
        id = id,
        roomId = roomId,
        category = category,
        pinned = pinned,
        title = title,
        content = content,
        images = images,
        placeName = placeName,
        latitude = latitude,
        longitude = longitude,
        startAt = startAt,
        endAt = endAt,
        extras = extras?.toDto(),
    )
}
