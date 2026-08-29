package uk.co.nekosune.presence

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONObject
import uk.co.nekosune.presence.databinding.ActivityMainBinding
import java.util.concurrent.Executors

class MainActivity : AppCompatActivity() {
    private lateinit var b: ActivityMainBinding
    private val http = OkHttpClient()
    private val worker = Executors.newSingleThreadExecutor()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityMainBinding.inflate(layoutInflater)
        setContentView(b.root)
        PresenceProtocol.ensureKey()
        b.host.setText(PairStore.host(this) ?: "")
        refreshStatus()
        requestRuntimePermissions()

        b.pairButton.setOnClickListener { pair() }
        b.startButton.setOnClickListener {
            if (!PairStore.paired(this)) {
                b.status.text = "Pair this phone first."
                return@setOnClickListener
            }
            ContextCompat.startForegroundService(this, Intent(this, PresenceService::class.java))
            b.status.text = "Presence service started"
        }
        b.stopButton.setOnClickListener {
            stopService(Intent(this, PresenceService::class.java))
            b.status.text = "Presence service stopped"
        }
    }

    private fun pair() {
        val host = b.host.text.toString().trim()
        val code = b.code.text.toString().trim()
        if (!NetworkGuard.isPrivateHost(host) || code.length != 6) {
            b.status.text = "Enter a private LAN IP and the 6-digit code from the PC."
            return
        }
        b.status.text = "Pairing…"
        worker.execute {
            try {
                val body = JSONObject()
                    .put("code", code)
                    .put("phone_public_key", PresenceProtocol.publicKeyRawHex())
                    .toString().toRequestBody("application/json".toMediaType())
                val req = Request.Builder().url("http://$host:45873/v1/pair").post(body).build()
                http.newCall(req).execute().use { res ->
                    if (!res.isSuccessful) error("PC returned HTTP ${res.code}")
                    val json = JSONObject(res.body?.string() ?: error("empty response"))
                    PairStore.save(this, host, json.getString("pair_id"), json.getString("computer_public_key"))
                }
                runOnUiThread {
                    b.status.text = "Paired successfully. Start the presence key."
                    refreshStatus()
                }
            } catch (e: Exception) {
                runOnUiThread { b.status.text = "Pairing failed: ${e.message}" }
            }
        }
    }

    private fun refreshStatus() {
        val paired = PairStore.paired(this)
        b.status.text = if (paired) "Paired to ${PairStore.host(this)}" else "Not paired"
        b.keyInfo.text = "Phone key: ${PresenceProtocol.publicKeyRawHex()}\n" +
            (PairStore.computerKey(this)?.let { "PC key: $it" } ?: "")
    }

    private fun requestRuntimePermissions() {
        val wanted = mutableListOf<String>()
        if (Build.VERSION.SDK_INT >= 33) {
            if (checkSelfPermission(Manifest.permission.NEARBY_WIFI_DEVICES) != PackageManager.PERMISSION_GRANTED)
                wanted += Manifest.permission.NEARBY_WIFI_DEVICES
            if (checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED)
                wanted += Manifest.permission.POST_NOTIFICATIONS
        }
        if (wanted.isNotEmpty()) ActivityCompat.requestPermissions(this, wanted.toTypedArray(), 100)
    }
}
