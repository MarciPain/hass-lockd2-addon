let appConfig = { locks: [] };
let haEntities = [];

document.addEventListener("DOMContentLoaded", async () => {
    await fetchHAEntities();
    await fetchConfig();
});

async function fetchConfig() {
    try {
        const res = await fetch('/api/config');
        if (res.ok) {
            appConfig = await res.json();
            if (!appConfig.locks) appConfig.locks = [];
            populateMainForm();
            renderEntityList();
        }
    } catch (e) {
        console.error("Failed to load config", e);
    }
}

async function fetchHAEntities() {
    try {
        const res = await fetch('/api/ha/entities');
        if (res.ok) {
            haEntities = await res.json();
            populateDropdowns();
        }
    } catch (e) {
        console.error("Failed to load HA entities", e);
    }
}

function populateMainForm() {
    document.getElementById('mqttHost').value = appConfig.MqttHost || '';
    document.getElementById('mqttPort').value = appConfig.MqttPort || 1883;
    document.getElementById('mqttUser').value = appConfig.MqttUser || '';
    document.getElementById('mqttPass').value = appConfig.MqttPass || '';
    document.getElementById('mqttSSL').checked = appConfig.MqttSSL || false;
}

async function saveConfig() {
    appConfig.MqttHost = document.getElementById('mqttHost').value;
    appConfig.MqttPort = parseInt(document.getElementById('mqttPort').value);
    appConfig.MqttUser = document.getElementById('mqttUser').value;
    appConfig.MqttPass = document.getElementById('mqttPass').value;
    appConfig.MqttSSL = document.getElementById('mqttSSL').checked;

    try {
        const res = await fetch('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(appConfig)
        });
        if (res.ok) {
            alert("Beállítások sikeresen mentve!");
        } else {
            const errText = await res.text();
            alert("Hiba mentés közben: " + errText);
        }
    } catch (e) {
        alert("Hiba a hálózaton: " + e.message);
    }
}

async function testMQTT() {
    const testConfig = {
        MqttHost: document.getElementById('mqttHost').value,
        MqttPort: parseInt(document.getElementById('mqttPort').value),
        MqttUser: document.getElementById('mqttUser').value,
        MqttPass: document.getElementById('mqttPass').value,
        MqttSSL: document.getElementById('mqttSSL').checked
    };

    if (!testConfig.MqttHost) {
        alert("Kérlek adj meg egy Host címet!");
        return;
    }

    try {
        const res = await fetch('/api/mqtt/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(testConfig)
        });

        const data = await res.json();
        if (res.ok && data.status === 'ok') {
            alert("✅ Sikeres csatlakozás az MQTT szerverhez!");
        } else {
            alert("❌ Sikertelen csatlakozás: " + (data.message || "Ismeretlen hiba"));
        }
    } catch (e) {
        alert("Hiba a tesztelés során: " + e.message);
    }
}

function renderEntityList() {
    const list = document.getElementById('entityList');
    list.innerHTML = '';

    if (appConfig.locks.length === 0) {
        list.innerHTML = '<p>Még nincs hozzáadva eszköz.</p>';
        return;
    }

    appConfig.locks.forEach((lock, idx) => {
        const div = document.createElement('div');
        div.className = `entity-list-item ${lock.enabled ? '' : 'disabled'}`;

        let subText = `Entitás: ${lock.entity_id} | Topic: /locks/${lock.topic_suffix}`;
        if (lock.mode === 'pulse') subText += ` | Mód: Pulse (${lock.pulse_duration}s)`;

        div.innerHTML = `
            <div>
                <h4 style="margin:0 0 5px 0">${lock.name} ${lock.enabled ? '🟢' : '🔴'}</h4>
                <small style="opacity: 0.8">${subText}</small>
            </div>
            <div>
                <button onclick="toggleEntity(${idx})">${lock.enabled ? 'Kikapcsolás' : 'Bekapcsolás'}</button>
                <button onclick="editEntity(${idx})">Szerkesztés</button>
                <button class="danger" onclick="deleteEntity(${idx})">Törlés</button>
            </div>
        `;
        list.appendChild(div);
    });
}

function populateDropdowns() {
    const mainSelect = document.getElementById('editHaEntity');
    const batterySelect = document.getElementById('editBatteryEntity');

    mainSelect.innerHTML = '<option value="">-- Válassz entitást --</option>';
    batterySelect.innerHTML = '<option value="">-- Válassz entitást --</option>';

    haEntities.forEach(e => {
        const name = e.attributes.friendly_name ? `${e.attributes.friendly_name} (${e.entity_id})` : e.entity_id;
        const opt = new Option(name, e.entity_id);

        if (e.entity_id.startsWith('sensor.')) {
            batterySelect.add(opt);
        } else {
            mainSelect.add(opt);
        }
    });
}

function toggleBattery() {
    const isChecked = document.getElementById('hasBattery').checked;
    document.getElementById('batteryContainer').classList.toggle('hidden', !isChecked);
}

function checkEntityType() {
    const val = document.getElementById('editHaEntity').value;
    const modeContainer = document.getElementById('modeContainer');

    // Ha switch-et választott, mutassuk a mode választót (Toggle/Pulse)
    if (val.startsWith('switch.')) {
        modeContainer.classList.remove('hidden');
        checkModeType();
    } else {
        modeContainer.classList.add('hidden');
        document.getElementById('pulseContainer').classList.add('hidden');
    }
}

function checkModeType() {
    const mode = document.getElementById('editMode').value;
    const pulseContainer = document.getElementById('pulseContainer');

    if (mode === 'pulse') {
        pulseContainer.classList.remove('hidden');
    } else {
        pulseContainer.classList.add('hidden');
    }
}

function showAddModal() {
    document.getElementById('modalTitle').innerText = 'Új Entitás';
    document.getElementById('editId').value = '';
    document.getElementById('editName').value = '';
    document.getElementById('editHaEntity').value = '';
    document.getElementById('hasBattery').checked = false;
    document.getElementById('editBatteryEntity').value = '';
    document.getElementById('editTopicSuffix').value = '';
    document.getElementById('editMode').value = 'toggle';
    document.getElementById('editPulseDuration').value = '2';

    toggleBattery();
    checkEntityType();

    document.getElementById('editModal').classList.remove('hidden');
    document.getElementById('modalBackdrop').classList.remove('hidden');
}

function closeModal() {
    document.getElementById('editModal').classList.add('hidden');
    document.getElementById('modalBackdrop').classList.add('hidden');
}

function saveEntity() {
    const idVal = document.getElementById('editId').value;
    const name = document.getElementById('editName').value;
    const entityId = document.getElementById('editHaEntity').value;
    const suffix = document.getElementById('editTopicSuffix').value;
    const hasBattery = document.getElementById('hasBattery').checked;
    const batteryId = document.getElementById('editBatteryEntity').value;
    const mode = document.getElementById('editMode').value;
    const pulseLength = parseInt(document.getElementById('editPulseDuration').value);

    if (!name || !entityId || !suffix) {
        alert("Kérlek töltsd ki az összes kötelező mezőt!");
        return;
    }

    const lockData = {
        id: idVal ? idVal : Date.now().toString(),
        name: name,
        entity_id: entityId,
        topic_suffix: suffix,
        battery_entity: hasBattery ? batteryId : "",
        enabled: true,
        mode: mode,
        pulse_duration: pulseLength
    };

    if (idVal) {
        // Edit 
        const extList = appConfig.locks;
        const index = extList.findIndex(l => l.id === idVal);
        if (index > -1) {
            lockData.enabled = extList[index].enabled; // preserve state
            extList[index] = lockData;
        }
    } else {
        // Add new
        appConfig.locks.push(lockData);
    }

    closeModal();
    saveConfig(); // API hívás + backend felé megy
}

function toggleEntity(idx) {
    appConfig.locks[idx].enabled = !appConfig.locks[idx].enabled;
    saveConfig();
}

function deleteEntity(idx) {
    if (confirm("Biztosan törölni szeretnéd az entitást?")) {
        appConfig.locks.splice(idx, 1);
        saveConfig();
    }
}

function editEntity(idx) {
    const lock = appConfig.locks[idx];
    document.getElementById('modalTitle').innerText = 'Entitás Szerkesztése';
    document.getElementById('editId').value = lock.id;
    document.getElementById('editName').value = lock.name;
    document.getElementById('editHaEntity').value = lock.entity_id;
    document.getElementById('editTopicSuffix').value = lock.topic_suffix;

    document.getElementById('hasBattery').checked = !!lock.battery_entity;
    document.getElementById('editBatteryEntity').value = lock.battery_entity || '';

    document.getElementById('editMode').value = lock.mode || 'toggle';
    document.getElementById('editPulseDuration').value = lock.pulse_duration || 2;

    toggleBattery();
    checkEntityType();

    document.getElementById('editModal').classList.remove('hidden');
    document.getElementById('modalBackdrop').classList.remove('hidden');
}


