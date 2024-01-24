import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
  stages: [
    { duration: '10s', target: 10 }, // Simulate 10 users over a period of 10 seconds
  ],
};

export default function () {
  http.get('http://localhost:8080/test');
  sleep(1);
}