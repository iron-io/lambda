package lambdatest;

import com.amazonaws.services.lambda.runtime.Context;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.Map;

public class Hello {
    public static void myHandlerPOJO(RequestClass request, Context context){
        System.out.println(String.format("Hello %s %s" , request.firstName, request.lastName));
    }
}
