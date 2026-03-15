import Axios from 'axios';

Axios.defaults.timeout = 180000;

interface IConfig {
  baseURL: string,
  headers: Record<string, any>,
}

const client = Axios.create({
  timeout: 180000,
  withCredentials: true,
});

client.interceptors.request.use((config: any) => {
  const token = localStorage.getItem('wm_token') || '';
  const headers: Record<string, any> = {
    ...(config.headers || {}),
    'cms-channel': 0,
  };

  if (token) {
    headers.Authorization = 'Bearer ' + token;
  }

  config.headers = headers;
  return config;
});

export default class Api {
  public static axios(_path: string, _data?: any, method: string = 'POST') {
    const config: IConfig = {
      baseURL: '/',
      headers: {
        'cms-channel': 0,
      },
    };

    if (method.toUpperCase() === 'GET') {
      return client.request({
        method,
        baseURL: config.baseURL,
        url: _path,
        params: _data,
        headers: config.headers,
      });
    }

    const formData = new FormData();
    if (_data && typeof _data === 'object') {
      for (const key in _data) {
        if (_data.hasOwnProperty(key) && _data[key] !== undefined && _data[key] !== null) {
          formData.append(key, _data[key]);
        }
      }
    }

    return client.request({
      method,
      baseURL: config.baseURL,
      url: _path,
      data: formData,
      headers: config.headers,
    });
  }
}
