import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { SharedArray } from 'k6/data';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp-up to 10 users
    { duration: '1m', target: 30 },   // Ramp-up to 30 users
    { duration: '2m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 0 },   // Ramp-down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

// This array should be populated with order IDs from a previous test run
// or you can use the complete-flow.js script instead
const orderIds = new SharedArray('order_ids', function () {
  return JSON.parse(open('./order-ids.json') || '[]');
});

export default function () {
  if (orderIds.length === 0) {
    console.error('No order IDs available. Please run create-orders.js first or use complete-flow.js');
    return;
  }

  const orderId = orderIds[Math.floor(Math.random() * orderIds.length)];

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.patch(`${BASE_URL}/api/v1/orders/${orderId}`, null, params);

  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response has id': (r) => JSON.parse(r.body).id !== undefined,
    'response has status': (r) => JSON.parse(r.body).status !== undefined,
  });

  errorRate.add(!success);

  sleep(1);
}
