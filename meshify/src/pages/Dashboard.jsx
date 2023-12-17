import React, { useState, useEffect } from 'react'
import { FaPlus, FaUniversalAccess } from 'react-icons/fa';
import Header from "../../src/Components/Header/Header";
import NavigationDrawer from '../Components/NavDrawer/NatvigationDrawer';
import Footer from '../Components/Footer/Footer';
import axios from "axios"
import {  RotateSpinner } from 'react-spinners-kit';
import { MdWarningAmber } from 'react-icons/md';


export default function Dashboard() {
  const [workloads, setWorkloads] = useState([]);
  const [error, setError] = useState(null);
  const [clusters, setClusters] = useState([]);
  const [loading, setLoading] = useState(false);



  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await axios.get("http://localhost:8080/api/kube/workloads");
        setWorkloads(response.data);
        setLoading(false);
      } catch (error) {
        setLoading(false);
       setError(error)
      }
    };
    

    fetchData();
  }, []);

  useEffect(() => {
    async function fetchClusterInfo() {
      try {
        setLoading(true);
        const response = await axios.get("http://localhost:8080/api/get/cluster");
        setClusters(response.data.clusters);
        setLoading(false);
      } catch (error) {
        setLoading(true);
        console.error(error);
      }
    }

    fetchClusterInfo();
  }, []);



  if (error) {
    return (
      <>
       <div className="bg-[#E8EFF6] text-red-800 py-4 px-6 rounded-md mt-4 ml-10 mr-10 mt-10">
          <div className="flex items-center">
            <span className="mr-2"><MdWarningAmber className="text-2xl" /></span>
            <span className="font-bold">Error:</span>
          </div>
          <p className="mt-2">
            Oops! Something went wrong. Don't panic, but we encountered an error.
          </p>
          <p className="mt-2 font-bold">Error Details:</p>
          <div className="text-red-700 py-2 px-4 rounded-md mt-2">
            {error.message}
          </div>
          <p className="mt-4">Please check your server is working or contact support for assistance.</p>
          <p className="mt-2">In the meantime, why not take a deep breath and enjoy this funky message:</p>
          <p className="mt-4 text-2xl font-bold text-">
            Oops! The gremlins are at it again! 🚀
          </p>
        </div>
      </>
     

    );
  }

  
  return (
    <>
      <Header />
      <div className="flex">
        <NavigationDrawer />

        <div className="float-right flex-1 inline-flex ml-4">
          <div className="bg-white shadow-md mx-4 my-8 px-4 py-3 rounded-md w-[700px]">
            {loading && (
              <div className="fixed bottom-12 right-9">
                <RotateSpinner size={50} color="#111827" />
              </div>
            )}
            <h3 className="text-xl font-medium mb-2">Kubernetes Workloads</h3>
            <div className="overflow-y-auto max-h-100 mt-10">
              <ul className="border border-gray-300 rounded-md p-4 ">
                {/* ... (existing code for Kubernetes Workloads) */}
              </ul>
            </div>
          </div>

          <div className="flex flex-col w-3/4 p-4 space-y-4 mt-7">
            <div className="bg-white shadow-md mx-2 p-4 h-[130px] w-[400px]">
              {/* ... (existing code for Connection Status) */}
            </div>

            <div className="bg-white shadow-md mx-2 p-4 h-[250px] w-[400px]">
              {/* ... (existing code for Service mesh adapters) */}
            </div>

            <div className="bg-white shadow-md mx-2 p-4 h-[170px] w-[400px]">
              {/* ... (existing code for Performance Graph) */}
            </div>
          </div>
        </div>
      </div>

      <Footer />
    </>
  )
}