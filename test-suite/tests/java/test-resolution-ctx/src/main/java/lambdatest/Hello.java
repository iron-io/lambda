package lambdatest;

import com.amazonaws.services.lambda.runtime.Context;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class Hello {
    public static void myHandler(int payload, Context context) {
        System.out.println(String.format("Amount of pain Java causes is %d", payload));
    }

    public static void myHandler(int payload) {
        System.out.println(String.format("Should not get executed %d", 1/0));
    }
}
