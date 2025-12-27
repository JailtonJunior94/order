import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp-up to 10 users
    { duration: '1m', target: 50 },   // Ramp-up to 50 users
    { duration: '2m', target: 100 },  // Stay at 100 users
    { duration: '30s', target: 0 },   // Ramp-down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% of requests should be below 500ms, 99% below 1s
    http_req_failed: ['rate<0.05'], // Error rate should be less than 5%
    errors: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

function generateRandomClient() {
  const clients = [
    'client-001',
    'client-002',
    'client-003',
    'client-004',
    'client-005',
  ];
  return clients[Math.floor(Math.random() * clients.length)];
}

function generateRandomItems() {
  const products = [
    { name: 'Product A', price: 29.99 },
    { name: 'Product B', price: 49.99 },
    { name: 'Product C', price: 19.99 },
    { name: 'Product D', price: 99.99 },
    { name: 'Product E', price: 39.99 },
  ];

  const numItems = Math.floor(Math.random() * 3) + 1; // 1-3 items
  const items = [];

  for (let i = 0; i < numItems; i++) {
    const product = products[Math.floor(Math.random() * products.length)];
    items.push({
      product_name: product.name,
      price: product.price,
      quantity: Math.floor(Math.random() * 5) + 1, // 1-5 quantity
    });
  }

  return items;
}

export default function () {
  const payload = JSON.stringify({
    client_id: generateRandomClient(),
    items: generateRandomItems(),
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(`${BASE_URL}/api/v1/orders`, payload, params);

  const success = check(res, {
    'status is 201': (r) => r.status === 201,
    'response has id': (r) => JSON.parse(r.body).id !== undefined,
    'response has status': (r) => JSON.parse(r.body).status !== undefined,
  });

  errorRate.add(!success);

  sleep(1);
}
