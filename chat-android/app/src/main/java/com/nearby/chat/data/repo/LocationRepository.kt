package com.nearby.chat.data.repo

import com.nearby.chat.core.model.GeocodeResult
import com.nearby.chat.core.model.LatLng
import com.nearby.chat.core.model.Poi
import com.nearby.chat.core.model.RegionItem
import com.nearby.chat.core.model.SavedChatLocation
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.local.ChatLocationStore
import com.nearby.chat.data.location.AppLocationProvider
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.remote.api.RegionApi
import com.nearby.chat.data.remote.tianditu.TiandituApi
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class LocationRepository @Inject constructor(
    private val regionApi: RegionApi,
    private val tiandituApi: TiandituApi,
    private val chatLocationStore: ChatLocationStore,
    private val locationProvider: AppLocationProvider,
) {
    fun hasMapKey(): Boolean = tiandituApi.hasKey()

    suspend fun reverseGeocode(lat: Double, lng: Double): GeocodeResult = tiandituApi.reverseGeocode(lat, lng)

    suspend fun searchPoiNearby(keyword: String, centerLat: Double, centerLng: Double): List<Poi> = tiandituApi.searchPoiNearby(keyword, centerLat, centerLng)

    suspend fun searchPoiInRegion(keyword: String, tiandituGb: String): List<Poi> = tiandituApi.searchPoiInRegion(keyword, tiandituGb)

    suspend fun regionCenter(tiandituGb: String, regionName: String): LatLng? = tiandituApi.getRegionCenter(tiandituGb, regionName)

    suspend fun regionChildren(parentCode: String = ""): List<RegionItem> = regionApi.children(parentCode).unwrap().list.map { it.toDomain() }

    suspend fun searchRegions(keyword: String, level: Int? = null, limit: Int? = null): List<RegionItem> = regionApi.search(keyword = keyword, level = level, limit = limit).unwrap().list.map { it.toDomain() }

    suspend fun savedLocation(): SavedChatLocation? = chatLocationStore.load()

    suspend fun saveLocation(lat: Double, lng: Double, placeName: String, browsing: Boolean) = chatLocationStore.save(lat, lng, placeName, browsing)

    fun peekCachedLocation(): LatLng? = locationProvider.peekCachedLocation()?.let { LatLng(it.latitude, it.longitude) }

    suspend fun currentLocation(timeoutMs: Long = 6_000): LatLng? = locationProvider.getLocation(timeoutMs)?.let { LatLng(it.latitude, it.longitude) }

    suspend fun refreshLocation(): LatLng? = locationProvider.refreshLocation()?.let { LatLng(it.latitude, it.longitude) }
}
