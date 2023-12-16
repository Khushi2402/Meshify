import React from "react";
import logo from '../../logo.svg';

export default function Example() {
  return (
    <div className="bg-white min-h-screen flex items-center justify-center">
      <div className="bg-gray-200 p-8 sm:w-full sm:max-w-2/3 lg:w-1/2 mx-auto">
        <div className="sm:mx-auto sm:w-full sm:max-w-sm">
          <span>
            <img
              className="mx-auto h-auto w-auto"
              src={logo}
              alt="Your Company"
            />

            <h1 className="company text-2xl text-indigo-600">MESHIFY</h1>
          </span>

          <h2 className="headings mt-10 text-center text-xl font-bold leading-9 tracking-tight text-gray-900">
            Select a Provider
          </h2>
        </div>

        <div className="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
          <form className="space-y-6" action="#" method="POST">
            <div>
              <button
                type="submit"
                className="buttons flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
              >
                Github
              </button>

              <br />
              <button
                type="submit"
                className="buttons flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
              >
                None
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
