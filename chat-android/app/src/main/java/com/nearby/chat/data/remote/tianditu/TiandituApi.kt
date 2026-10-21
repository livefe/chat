package com.nearby.chat.data.remote.tianditu

import android.content.Context
import com.nearby.chat.BuildConfig
import com.nearby.chat.R
import com.nearby.chat.core.model.GeocodeResult
import com.nearby.chat.core.model.LatLng
import com.nearby.chat.core.model.Poi
import com.nearby.chat.di.TiandituHttpClient
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlinx.serialization.json.put
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.Request
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TiandituApi @Inject constructor(
    @TiandituHttpClient private val client: OkHttpClient,
    private val json: Json,
    @ApplicationContext private val context: Context,
) {
    fun hasKey(): Boolean = BuildConfig.TIANDITU_TK.isNotBlank()

    suspend fun reverseGeocode(lat: Double, lng: Double): GeocodeResult = withContext(Dispatchers.IO) {
        val postStr = buildJsonObject {
            put("lon", lng)
            put("lat", lat)
            put("ver", 1)
        }.toString()
        val root = get("/geocoder", mapOf("postStr" to postStr, "type" to "geocode"))
        val result = root["result"]?.jsonObject
        val formatted = result?.get("formatted_address")?.jsonPrimitive?.contentOrNull.orEmpty()
        val comp = result?.get("addressComponent")?.jsonObject
        GeocodeResult(
            placeName = formatPlaceName(comp, formatted),
            formattedAddress = formatted,
            province = comp?.get("province")?.jsonPrimitive?.contentOrNull.orEmpty(),
            city = comp?.get("city")?.jsonPrimitive?.contentOrNull.orEmpty(),
            county = comp?.get("county")?.jsonPrimitive?.contentOrNull.orEmpty(),
        )
    }

    suspend fun searchPoiNearby(
        keyword: String,
        centerLat: Double,
        centerLng: Double,
        count: Int = 20,
    ): List<Poi> = withContext(Dispatchers.IO) {
        val keyWord = keyword.trim()
        if (keyWord.isEmpty()) return@withContext emptyList()
        val postStr = buildJsonObject {
            put("keyWord", keyWord)
            put("level", 16)
            put("queryRadius", 5000)
            put("pointLonlat", "$centerLng,$centerLat")
            put("queryType", 3)
            put("start", 0)
            put("count", count)
        }.toString()
        val root = get("/v2/search", mapOf("postStr" to postStr, "type" to "query"))
        parsePois(root["pois"] as? JsonArray ?: root["pois"]?.jsonArray)
    }

    suspend fun searchPoiInRegion(
        keyword: String,
        tiandituGb: String,
        count: Int = 20,
    ): List<Poi> = withContext(Dispatchers.IO) {
        val keyWord = keyword.trim()
        if (keyWord.isEmpty() || tiandituGb.isBlank()) return@withContext emptyList()
        val postStr = buildJsonObject {
            put("keyWord", keyWord)
            put("queryType", 12)
            put("specify", tiandituGb)
            put("start", 0)
            put("count", count)
            put("show", 2)
        }.toString()
        val root = get("/v2/search", mapOf("postStr" to postStr, "type" to "query"))
        parsePois(root["pois"] as? JsonArray ?: root["pois"]?.jsonArray)
    }

    suspend fun getRegionCenter(tiandituGb: String, regionName: String = ""): LatLng? = withContext(Dispatchers.IO) {
        if (tiandituGb.isBlank()) return@withContext null
        runCatching {
            val root = get(
                "/v2/administrative",
                mapOf(
                    "keyword" to tiandituGb,
                    "childLevel" to "0",
                    "extensions" to "true",
                ),
            )
            val districts = root["data"]?.jsonObject?.get("district")?.jsonArray
                ?: root["district"]?.jsonArray
            val first = districts?.firstOrNull()?.jsonObject
            parseLngLat(first?.get("center")?.jsonPrimitive?.contentOrNull)
                ?: centerFromBound(first?.get("bound")?.jsonPrimitive?.contentOrNull)
        }.getOrNull() ?: run {
            if (regionName.isBlank()) return@withContext null
            searchPoiInRegion(regionName, tiandituGb, count = 1).firstOrNull()
                ?.let { LatLng(it.lat, it.lng) }
        }
    }

    private fun parseLngLat(raw: String?): LatLng? {
        if (raw.isNullOrBlank()) return null
        val parts = raw.split(",")
        if (parts.size < 2) return null
        val lng = parts[0].toDoubleOrNull() ?: return null
        val lat = parts[1].toDoubleOrNull() ?: return null
        return LatLng(lat, lng)
    }

    private fun centerFromBound(bound: String?): LatLng? {
        if (bound.isNullOrBlank()) return null
        val nums = bound.split(",").mapNotNull { it.trim().toDoubleOrNull() }
        if (nums.size < 4) return null
        val lng = (nums[0] + nums[2]) / 2
        val lat = (nums[1] + nums[3]) / 2
        return LatLng(lat, lng)
    }

    private fun parsePois(arr: JsonArray?): List<Poi> {
        if (arr == null) return emptyList()
        val unknown = context.getString(R.string.map_unknown_place)
        return arr.mapNotNull { el ->
            val obj = el.jsonObject
            val lonlat = obj["lonlat"]?.jsonPrimitive?.contentOrNull.orEmpty()
            val parts = lonlat.split(",")
            if (parts.size < 2) return@mapNotNull null
            val lng = parts[0].toDoubleOrNull() ?: return@mapNotNull null
            val lat = parts[1].toDoubleOrNull() ?: return@mapNotNull null
            Poi(
                name = obj["name"]?.jsonPrimitive?.contentOrNull?.ifBlank { unknown } ?: unknown,
                address = obj["address"]?.jsonPrimitive?.contentOrNull.orEmpty(),
                lat = lat,
                lng = lng,
            )
        }
    }

    private fun get(path: String, params: Map<String, String>): JsonObject {
        val tk = BuildConfig.TIANDITU_TK
        if (tk.isBlank()) error(context.getString(R.string.map_key_missing))
        val urlBuilder = "$API_BASE$path".toHttpUrl().newBuilder().addQueryParameter("tk", tk)
        params.forEach { (k, v) -> urlBuilder.addQueryParameter(k, v) }
        val req = Request.Builder().url(urlBuilder.build()).get().build()
        client.newCall(req).execute().use { resp ->
            if (!resp.isSuccessful) error(context.getString(R.string.map_request_failed, resp.code))
            val body = resp.body?.string().orEmpty()
            val root = json.parseToJsonElement(body).jsonObject
            ensureSuccess(root)
            return root
        }
    }

    private fun ensureSuccess(root: JsonObject) {
        val status = root["status"] ?: return
        when (status) {
            is JsonPrimitive -> {
                val text = status.contentOrNull.orEmpty()
                if (text.isEmpty() || text == "0" || text == "1000" || text == "200") return
                error(
                    root["msg"]?.jsonPrimitive?.contentOrNull
                        ?: root["remark"]?.jsonPrimitive?.contentOrNull
                        ?: context.getString(R.string.map_service_error),
                )
            }

            is JsonObject -> {
                val code = status["infocode"]?.jsonPrimitive?.contentOrNull?.toIntOrNull() ?: 0
                if (code == 0 || code == 1000) return
                error(status["cndesc"]?.jsonPrimitive?.contentOrNull ?: context.getString(R.string.map_service_error))
            }

            else -> Unit
        }
    }

    private fun formatPlaceName(comp: JsonObject?, formatted: String): String {
        val poi = comp?.get("poi")?.jsonPrimitive?.contentOrNull?.trim().orEmpty()
        if (poi.isNotEmpty()) return poi
        return formatted.trim()
    }

    private companion object {
        const val API_BASE = "https://api.tianditu.gov.cn"
    }
}
