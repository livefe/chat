package com.nearby.chat.data.repo

import com.nearby.chat.core.model.AdEvent
import com.nearby.chat.core.network.ensureSuccess
import com.nearby.chat.data.mapper.toRequest
import com.nearby.chat.data.remote.api.AdApi
import com.nearby.chat.data.remote.dto.AdEventsRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AdRepository @Inject constructor(
    private val adApi: AdApi,
) {
    suspend fun reportEvents(events: List<AdEvent>) {
        if (events.isEmpty()) return
        adApi.reportEvents(AdEventsRequest(events.map { it.toRequest() })).ensureSuccess()
    }
}
