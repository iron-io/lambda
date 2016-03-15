package lambdatest;

import com.amazonaws.services.lambda.runtime.Context;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class Hello {
    public static void myHandler(Context context) {
        System.out.println("This gets called.");

        String fn = context.getFunctionName();
        assert fn != null;
        assert fn.length() > 0;
        assert fn.contains("context");

        String vsn = context.getFunctionVersion();
        assert vsn != null;
        assert vsn.equals("$LATEST");

        String id = context.getAwsRequestId();
        assert id != null;
        assert id.length() > 0;

        int timeLeft = context.getRemainingTimeInMillis();
        assert timeLeft >= 0;
        try { Thread.currentThread().sleep(500); } catch(Exception e) {}
        int newTimeLeft = context.getRemainingTimeInMillis();
        assert newTimeLeft >= 0;
        assert newTimeLeft <= timeLeft;

        int mem = context.getMemoryLimitInMB();
        assert mem > 0;
    }
}
