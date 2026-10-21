package com.nearby.chat.data.network

import com.nearby.chat.BuildConfig
import com.nearby.chat.core.network.ClientHeaders
import com.nearby.chat.data.local.DeviceId
import okhttp3.Interceptor
import okhttp3.Response
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ClientContextInterceptor @Inject constructor(
    private val deviceId: DeviceId,
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request().newBuilder()
            .header(ClientHeaders.REQUEST_ID, UUID.randomUUID().toString())
            .header(ClientHeaders.CLIENT_PLATFORM, ClientHeaders.PLATFORM_ANDROID)
            .header(ClientHeaders.CLIENT_VERSION, BuildConfig.VERSION_NAME)
            .header(ClientHeaders.DEVICE_ID, deviceId.getBlocking())
            .build()
        return chain.proceed(request)
    }
}
