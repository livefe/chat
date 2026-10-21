package com.nearby.chat.data.repo

import com.nearby.chat.core.network.ensureSuccess
import com.nearby.chat.data.remote.api.ReportApi
import com.nearby.chat.data.remote.dto.SubmitReportRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ReportRepository @Inject constructor(
    private val reportApi: ReportApi,
) {
    suspend fun submit(
        targetType: Int,
        targetId: Long,
        reason: Int,
        content: String?,
        images: List<Long>?,
    ) {
        reportApi.submit(
            SubmitReportRequest(
                targetType = targetType,
                targetId = targetId,
                reason = reason,
                content = content,
                images = images,
            ),
        ).ensureSuccess()
    }
}
