package com.nearby.chat.data.repo

import android.os.Build
import com.nearby.chat.core.network.ensureSuccess
import com.nearby.chat.data.push.PushManager
import com.nearby.chat.data.remote.api.DeviceApi
import com.nearby.chat.data.remote.dto.DeviceRegisterRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class DeviceRepository @Inject constructor(
    private val deviceApi: DeviceApi,
    private val pushManager: PushManager,
) {
    suspend fun register(pushToken: String) {
        if (pushToken.isBlank()) return
        val req = DeviceRegisterRequest(
            vendor = pushManager.vendor.apiValue(),
            pushToken = pushToken,
            osVersion = Build.VERSION.RELEASE.orEmpty(),
        )
        deviceApi.register(req).ensureSuccess()
    }

    suspend fun unbind() {
        deviceApi.unbind().ensureSuccess()
    }
}
