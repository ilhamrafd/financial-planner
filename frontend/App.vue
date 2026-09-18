<template>
  <div class="mybca-container">
    <!-- Navbar / Header ala myBCA -->
    <header class="mybca-header shadow-sm">
      <div class="container d-flex justify-content-between align-items-center py-3">
        <div class="d-flex align-items-center gap-2">
          <div class="brand-logo">my<b>BCA</b></div>
        </div>
        <div class="d-flex align-items-center gap-3">
          <span class="user-greeting">Halo, <b>{{ currentUser.name || 'King' }}</b></span>
          <button @click="logout" class="btn btn-outline-danger btn-sm rounded-pill px-3">Keluar</button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="container my-4">
      <!-- Akun Selector -->
      <div class="row mb-4">
        <div class="col-md-4">
          <label class="form-label text-muted small font-weight-bold">Pilih Akun / Rekening:</label>
          <select v-model.number="userId" @change="fetchWallet" class="form-select rounded-3 shadow-none border-0 bg-white p-2 fw-bold text-primary">
            <option v-for="user in users" :value="user.id" :key="user.id">
              {{ user.name }} (ID: {{ user.id }})
            </option>
          </select>
        </div>
      </div>

      <div class="row g-4">
        <!-- Kolom Kiri: Ringkasan Rekening & Saldo -->
        <div class="col-lg-8">
          <!-- Card Rekening Utama -->
          <div class="card bca-card text-white p-4 mb-4 shadow-sm border-0 rounded-4">
            <div class="d-flex justify-content-between align-items-start mb-3">
              <div>
                <span class="badge bg-light text-dark mb-2 px-2 py-1 rounded-pill small">Tahapan BCA</span>
                <h6 class="text-white-50 m-0">No. Rekening</h6>
                <h5 class="fw-bold tracking-wider">0281 999 888</h5>
              </div>
              <div class="text-end">
                <span class="badge bg-success bg-opacity-75 px-2 py-1 rounded-pill small">Aktif</span>
              </div>
            </div>
            <div class="mt-3">
              <span class="text-white-50 small">Total Saldo Tersedia</span>
              <div class="d-flex align-items-center gap-3 mt-1">
                <h2 class="fw-bold mb-0">
                  <span v-if="showBalance">Rp {{ balance.toLocaleString() }}</span>
                  <span v-else>Rp ••••••••</span>
                </h2>
                <button @click="showBalance = !showBalance" class="btn btn-sm btn-link text-white text-decoration-none p-0">
                  <i :class="showBalance ? 'bi bi-eye-slash' : 'bi bi-eye'"></i> {{ showBalance ? 'Sembunyikan' : 'Tampilkan' }}
                </button>
              </div>
            </div>
          </div>

          <!-- Quick Actions (Menu Cepat ala myBCA) -->
          <div class="card p-4 shadow-sm border-0 rounded-4 mb-4 bg-white">
            <h5 class="fw-bold mb-3 text-secondary small text-uppercase">Menu Utama</h5>
            <div class="row text-center g-3">
              <div class="col-3">
                <div @click="activeTab = 'deposit'" class="quick-menu-item p-3 rounded-4" :class="{ active: activeTab === 'deposit' }">
                  <div class="icon-circle bg-primary bg-opacity-10 text-primary mx-auto mb-2">📥</div>
                  <span class="small fw-bold text-dark">Deposit</span>
                </div>
              </div>
              <div class="col-3">
                <div @click="activeTab = 'withdraw'" class="quick-menu-item p-3 rounded-4" :class="{ active: activeTab === 'withdraw' }">
                  <div class="icon-circle bg-danger bg-opacity-10 text-danger mx-auto mb-2">📤</div>
                  <span class="small fw-bold text-dark">Tarik Tunai</span>
                </div>
              </div>
              <div class="col-3">
                <div @click="activeTab = 'transfer'" class="quick-menu-item p-3 rounded-4" :class="{ active: activeTab === 'transfer' }">
                  <div class="icon-circle bg-success bg-opacity-10 text-success mx-auto mb-2">🔄</div>
                  <span class="small fw-bold text-dark">Transfer</span>
                </div>
              </div>
              <div class="col-3">
                <div @click="activeTab = 'history'" class="quick-menu-item p-3 rounded-4" :class="{ active: activeTab === 'history' }">
                  <div class="icon-circle bg-warning bg-opacity-10 text-warning mx-auto mb-2">📋</div>
                  <span class="small fw-bold text-dark">Mutasi</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Kolom Kanan: Form Aksi Dinamis -->
        <div class="col-lg-4">
          <div class="card p-4 shadow-sm border-0 rounded-4 bg-white">
            <h5 class="fw-bold mb-3 text-dark">
              <span v-if="activeTab === 'deposit'">Top Up Saldo</span>
              <span v-else-if="activeTab === 'withdraw'">Tarik Saldo</span>
              <span v-else-if="activeTab === 'transfer'">Transfer Antar Rekening</span>
              <span v-else>Informasi Mutasi</span>
            </h5>

            <!-- Form Deposit / Withdraw -->
            <div v-if="activeTab === 'deposit' || activeTab === 'withdraw'">
              <div class="mb-3">
                <label class="form-label small text-muted">Nominal (Rp)</label>
                <div class="input-group">
                  <span class="input-group-text bg-light border-end-0">Rp</span>
                  <input type="number" v-model="amount" class="form-control border-start-0 bg-light shadow-none" placeholder="0" />
                </div>
              </div>
              <div class="mb-3">
                <label class="form-label small text-muted">Keterangan / Berita</label>
                <input type="text" v-model="description" class="form-control bg-light shadow-none" :placeholder="activeTab === 'deposit' ? 'Top Up myBCA' : 'Tarik Tunai'" />
              </div>
              <button @click="doTransaction(activeTab.toUpperCase())" class="btn btn-primary w-100 rounded-pill py-2 fw-bold shadow-sm" :class="activeTab === 'deposit' ? 'btn-primary' : 'btn-danger'">
                {{ activeTab === 'deposit' ? 'Konfirmasi Deposit' : 'Konfirmasi Penarikan' }}
              </button>
            </div>

            <!-- Form Transfer -->
            <div v-else-if="activeTab === 'transfer'">
              <div class="mb-3">
                <label class="form-label small text-muted">Rekening Tujuan (User ID)</label>
                <input type="number" v-model="targetUserId" class="form-control bg-light shadow-none" placeholder="Masukkan ID User Penerima" />
              </div>
              <div class="mb-3">
                <label class="form-label small text-muted">Nominal Transfer (Rp)</label>
                <div class="input-group">
                  <span class="input-group-text bg-light border-end-0">Rp</span>
                  <input type="number" v-model="amount" class="form-control border-start-0 bg-light shadow-none" placeholder="0" />
                </div>
              </div>
              <button @click="doTransfer" class="btn btn-success w-100 rounded-pill py-2 fw-bold shadow-sm">
                Transfer Sekarang
              </button>
            </div>

            <!-- Mutasi / Riwayat -->
            <div v-else>
              <p class="text-muted small">Fitur riwayat mutasi transaksi real-time.</p>
              <div class="list-group list-group-flush">
                <div v-for="(tx, idx) in transactions" :key="idx" class="list-group-item px-0 py-2 d-flex justify-content-between align-items-center">
                  <div>
                    <div class="small fw-bold text-dark">{{ tx.type }}</div>
                    <div class="text-muted" style="font-size: 11px;">{{ tx.desc }}</div>
                  </div>
                  <span class="fw-bold small" :class="tx.type === 'DEPOSIT' ? 'text-success' : 'text-danger'">
                    {{ tx.type === 'DEPOSIT' ? '+' : '-' }} Rp {{ tx.amount.toLocaleString() }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  data() {
    return {
      userId: 1,
      balance: 0,
      amount: 0,
      targetUserId: null,
      description: '',
      showBalance: true,
      activeTab: 'deposit',
      users: [],
      transactions: [
        { type: 'DEPOSIT', amount: 50000, desc: 'Top Up awal' }
      ]
    }
  },
  computed: {
    currentUser() {
      return this.users.find(u => u.id === this.userId) || {}
    }
  },
  mounted() {
    this.fetchUsers()
  },
  methods: {
    async fetchUsers() {
      try {
        const res = await axios.get('http://localhost:8080/api/users')
        this.users = res.data
        if (this.users.length > 0) {
          this.userId = this.users[0].id
          this.fetchWallet()
        }
      } catch (err) {
        console.error("Gagal mengambil data user", err)
      }
    },
    async fetchWallet() {
      try {
        const res = await axios.get(`http://localhost:8080/api/wallet/${this.userId}`)
        this.balance = res.data.balance
      } catch (err) {
        console.error("Gagal mengambil data wallet", err)
      }
    },
    async doTransaction(type) {
      if (this.amount <= 0) {
        alert("Masukkan nominal yang valid!")
        return
      }
      try {
        const res = await axios.post('http://localhost:8080/api/transaction', {
          user_id: parseInt(this.userId),
          amount: parseFloat(this.amount),
          type: type,
          description: this.description || `myBCA ${type}`
        })
        this.balance = res.data.new_balance
        this.transactions.unshift({
          type: type,
          amount: parseFloat(this.amount),
          desc: this.description || `myBCA ${type}`
        })
        this.amount = 0
        this.description = ''
        alert(`${type} Berhasil!`)
      } catch (err) {
        alert(err.response?.data || "Terjadi kesalahan transaksi")
      }
    },
    async doTransfer() {
      alert("Fitur transfer antar rekening akan segera hadir!")
    },
    logout() {
      alert("Berhasil keluar dari sesi myBCA.")
    }
  }
}
</script>

<style scoped>
.mybca-container {
  background-color: #f8f9fa;
  min-height: 100vh;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}
.mybca-header {
  background: #ffffff;
  border-bottom: 1px solid #eaeaea;
}
.brand-logo {
  font-size: 1.3rem;
  color: #0060af; /* Warna khas biru BCA */
  letter-spacing: -0.5px;
}
.bca-card {
  background: linear-gradient(135deg, #004b93 0%, #0060af 100%);
}
.quick-menu-item {
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fdfdfd;
  border: 1px solid #f0f0f0;
}
.quick-menu-item:hover, .quick-menu-item.active {
  background: #f0f4f9;
  border-color: #0060af;
  transform: translateY(-2px);
}
.icon-circle {
  width: 40px;
  height: 40px;
  display: flex;
  align-items-center-center;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 1.1rem;
}
</style>
