package com.nearby.chat.data.location

import android.annotation.SuppressLint
import android.content.Context
import android.location.Location
import android.location.LocationListener
import android.location.LocationManager
import android.os.Bundle
import android.os.Looper
import com.google.android.gms.common.ConnectionResult
import com.google.android.gms.common.GoogleApiAvailability
import com.google.android.gms.location.LocationServices
import com.google.android.gms.location.Priority
import com.google.android.gms.tasks.CancellationTokenSource
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.tasks.await
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeoutOrNull
import javax.inject.Inject
import javax.inject.Singleton
import kotlin.coroutines.resume

@Singleton
class AppLocationProvider @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    @Volatile
    private var memoryCache: Location? = null

    @SuppressLint("MissingPermission")
    fun peekCachedLocation(maxAgeMs: Long = CACHE_ACCEPT_MS): Location? {
        memoryCache?.takeIf { ageOf(it) <= maxAgeMs }?.let { return it }
        return systemLastKnown()?.also { remember(it) }
            ?.takeIf { ageOf(it) <= maxAgeMs }
    }

    @SuppressLint("MissingPermission")
    suspend fun getLocation(timeoutMs: Long = 6_000): Location? = withContext(Dispatchers.IO) {
        peekCachedLocation(CACHE_ACCEPT_MS)?.let { return@withContext it }

        val fresh = refreshLocation(timeoutMs)
        if (fresh != null) return@withContext fresh

        peekCachedLocation(STALE_ACCEPT_MS)
            ?: systemLastKnown()?.also { remember(it) }
    }

    @SuppressLint("MissingPermission")
    suspend fun refreshLocation(timeoutMs: Long = 6_000): Location? = withContext(Dispatchers.IO) {
        coroutineScope {
            val fusedJob = async {
                if (!hasGooglePlay()) {
                    null
                } else {
                    withTimeoutOrNull(timeoutMs.coerceAtMost(4_000)) { fusedLocationQuick() }
                }
            }
            val systemJob = async {
                withTimeoutOrNull(timeoutMs) { systemLocationQuick() }
            }
            val fused = fusedJob.await()
            if (fused != null) {
                systemJob.cancel()
                return@coroutineScope remember(fused)
            }
            systemJob.await()?.let { return@coroutineScope remember(it) }
            null
        }
    }

    private fun remember(loc: Location): Location {
        memoryCache = loc
        return loc
    }

    private fun ageOf(loc: Location): Long = System.currentTimeMillis() - loc.time

    private fun hasGooglePlay(): Boolean = GoogleApiAvailability.getInstance()
        .isGooglePlayServicesAvailable(context) == ConnectionResult.SUCCESS

    @SuppressLint("MissingPermission")
    private suspend fun fusedLocationQuick(): Location? {
        val client = LocationServices.getFusedLocationProviderClient(context)

        runCatching { client.lastLocation.await() }.getOrNull()
            ?.takeIf { ageOf(it) <= CACHE_ACCEPT_MS }
            ?.let { return it }

        return runCatching {
            client.getCurrentLocation(
                Priority.PRIORITY_BALANCED_POWER_ACCURACY,
                CancellationTokenSource().token,
            ).await()
        }.getOrNull()
            ?: runCatching { client.lastLocation.await() }.getOrNull()
    }

    @SuppressLint("MissingPermission")
    private fun systemLastKnown(): Location? {
        val lm = context.getSystemService(Context.LOCATION_SERVICE) as LocationManager
        return listOf(
            LocationManager.NETWORK_PROVIDER,
            LocationManager.GPS_PROVIDER,
            LocationManager.PASSIVE_PROVIDER,
        ).mapNotNull { provider ->
            runCatching {
                if (lm.isProviderEnabled(provider)) lm.getLastKnownLocation(provider) else null
            }.getOrNull()
        }.maxByOrNull { it.time }
    }

    @SuppressLint("MissingPermission")
    private suspend fun systemLocationQuick(): Location? {
        val last = systemLastKnown()
        if (last != null && ageOf(last) <= CACHE_ACCEPT_MS) return last

        val lm = context.getSystemService(Context.LOCATION_SERVICE) as LocationManager

        val network = withTimeoutOrNull(3_000) {
            requestSingleUpdate(lm, LocationManager.NETWORK_PROVIDER)
        }
        if (network != null) return network

        val gps = withTimeoutOrNull(3_000) {
            requestSingleUpdate(lm, LocationManager.GPS_PROVIDER)
        }
        return gps ?: last
    }

    @SuppressLint("MissingPermission")
    private suspend fun requestSingleUpdate(lm: LocationManager, provider: String): Location? {
        if (runCatching { !lm.isProviderEnabled(provider) }.getOrDefault(true)) return null
        return suspendCancellableCoroutine { cont ->
            val listener = object : LocationListener {
                override fun onLocationChanged(location: Location) {
                    runCatching { lm.removeUpdates(this) }
                    if (cont.isActive) cont.resume(location)
                }

                @Deprecated("Deprecated in Java")
                override fun onStatusChanged(provider: String?, status: Int, extras: Bundle?) {}
                override fun onProviderEnabled(provider: String) {}
                override fun onProviderDisabled(provider: String) {}
            }
            cont.invokeOnCancellation {
                runCatching { lm.removeUpdates(listener) }
            }
            runCatching {
                lm.requestLocationUpdates(provider, 0L, 0f, listener, Looper.getMainLooper())
            }.onFailure {
                if (cont.isActive) cont.resume(null)
            }
        }
    }

    companion object {

        const val CACHE_ACCEPT_MS = 5 * 60 * 1000L

        const val STALE_ACCEPT_MS = 2 * 60 * 60 * 1000L
    }
}
