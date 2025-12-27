import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter } from 'k6/metrics';

const errorRate = new Rate('errors');
const ordersCreated = new Counter('orders_created');
const ordersMarkedAsPaid = new Counter('orders_marked_as_paid');

export const options = {
  scenarios: {
    create_and_pay: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },  // Ramp-up to 10 users
        { duration: '1m', target: 50 },   // Ramp-up to 50 users
        { duration: '3m', target: 100 },  // Stay at 100 users
        { duration: '1m', target: 50 },   // Ramp-down to 50 users
        { duration: '30s', target: 0 },   // Ramp-down to 0 users
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.05'], // Error rate should be less than 5%
    errors: ['rate<0.05'],
    orders_created: ['count>0'],
    orders_marked_as_paid: ['count>0'],
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
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  // Step 1: Create an order
  const createPayload = JSON.stringify({
    client_id: generateRandomClient(),
    items: generateRandomItems(),
  });

  const createRes = http.post(`${BASE_URL}/api/v1/orders`, createPayload, params);

  const createSuccess = check(createRes, {
    'create: status is 201': (r) => r.status === 201,
    'create: response has id': (r) => {
      try {
        return JSON.parse(r.body).id !== undefined;
      } catch (e) {
        return false;
      }
    },
    'create: response has status': (r) => {
      try {
        return JSON.parse(r.body).status !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (!createSuccess) {
    errorRate.add(1);
    sleep(1);
    return;
  }

  ordersCreated.add(1);

  let orderId;
  try {
    const createBody = JSON.parse(createRes.body);
    orderId = createBody.id;
  } catch (e) {
    console.error('Failed to parse create order response:', e);
    errorRate.add(1);
    sleep(1);
    return;
  }

  // Wait a bit before marking as paid (simulating real user behavior)
  sleep(Math.random() * 2 + 1); // Random sleep between 1-3 seconds

  // Step 2: Mark the order as paid
  const markAsPaidRes = http.patch(`${BASE_URL}/api/v1/orders/${orderId}`, null, params);

  const markAsPaidSuccess = check(markAsPaidRes, {
    'mark_as_paid: status is 200': (r) => r.status === 200,
    'mark_as_paid: response has id': (r) => {
      try {
        return JSON.parse(r.body).id !== undefined;
      } catch (e) {
        return false;
      }
    },
    'mark_as_paid: response has status': (r) => {
      try {
        return JSON.parse(r.body).status !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (markAsPaidSuccess) {
    ordersMarkedAsPaid.add(1);
  } else {
    errorRate.add(1);
  }

  sleep(1);
}
