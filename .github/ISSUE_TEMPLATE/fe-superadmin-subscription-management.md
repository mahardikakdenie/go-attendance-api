untuk FE ketika Create Plans kenapa error 400

URL: http://localhost:3001/api/v1/superadmin/plans POST

rESPONSE BE
{
    "success": false,
    "meta": {
        "message": "Failed to create plan",
        "code": 500,
        "status": "error"
    },
    "data": "ERROR: column \"features\" is of type json but expression is of type record (SQLSTATE 42804)"
}BISA FIX 
