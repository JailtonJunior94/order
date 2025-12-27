import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter } from 'k6/metrics';

const errorRate = new Rate('errors');
const ordersCreated = new Counter('orders_created');
const ordersMarkedAsPaid = new Counter('orders_marked_as_paid');

export const options = {
  scenarios: {
    soak: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 20 },   // Ramp-up to 20 users
        { duration: '30m', target: 20 },  // Stay at 20 users for 30 minutes
        { duration: '2m', target: 0 },    // Ramp-down to 0 users
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'], // Very low error rate for soak test
    errors: ['rate<0.01'],
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

  const numItems = Math.floor(Math.random() * 3) + 1;
  const items = [];

  for (let i = 0; i < numItems; i++) {
    const product = products[Math.floor(Math.random() * products.length)];
    items.push({
      product_name: product.name,
      price: product.price,
      quantity: Math.floor(Math.random() * 5) + 1,
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

  // Create an order
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
  });

  if (!createSuccess) {
    errorRate.add(1);
    sleep(2);
    return;
  }

  ordersCreated.add(1);

  let orderId;
  try {
    const createBody = JSON.parse(createRes.body);
    orderId = createBody.id;
  } catch (e) {
    errorRate.add(1);
    sleep(2);
    return;
  }

  // Simulate realistic user behavior with random thinking time
  sleep(Math.random() * 3 + 2); // Random sleep between 2-5 seconds

  // Mark as paid
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
  });

  if (markAsPaidSuccess) {
    ordersMarkedAsPaid.add(1);
  } else {
    errorRate.add(1);
  }

  sleep(2);
}
