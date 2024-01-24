import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
  stages: [
    {
      duration: '10s',
      target: 10,
    },
  ],
};

export default function () {
  http.get('http://localhost:8080/test');
  sleep(1);
}