
import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  scenarios: {
    wallet_stress: {
      executor: 'constant-arrival-rate',
      rate: 1000, // Целевой RPS
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 1,
      maxVUs: 1,
    },
  },
  thresholds: {
    http_req_failed: ['rate<=0.00'], // Менее 1% ошибок
    http_req_duration: ['p(95)<500'], // 95% запросов быстрее 500ms
  },
};

const WALLET_ID = 'c9c5c5e0-7b3a-4e3a-9b3d-3d9b2e3d3d9b';
const BASE_URL = 'http://localhost:8080/api/v1';

export default function () {
  // Тестируем 50% депозитов и 50% проверок баланса
  if (Math.random() < 0.5) {
    const payload = JSON.stringify({
      walletId: WALLET_ID,
      operationType: Math.random() < 0.7 ? 'DEPOSIT' : 'WITHDRAW',
      amount: 100,
    });

    const res = http.post(`${BASE_URL}/wallet`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    check(res, {
      'POST status 200': (r) => r.status === 200,
    });
  } else {
    const res = http.get(`${BASE_URL}/wallets/${WALLET_ID}`);
    check(res, {
      'GET status 200': (r) => r.status === 200,
    });
  }
}